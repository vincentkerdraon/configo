package config_test

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/vincentkerdraon/configo/config"
	"github.com/vincentkerdraon/configo/config/param"
	"github.com/vincentkerdraon/configo/lock"
)

// When using declarative style. Define params one by one.
func Example_whenDeclarativeStyle() {
	type City string
	user := struct {
		Name string
		City City
		Age  int
	}{}

	//This is a minimal parameter declaration
	pName, err := param.New( //ID of the parameter. Also used to read flags and env vars.
		"Name",
		//this parse function defines what to do from the received string.
		//same Signature as Set() in std flag lib.
		func(s string) error { user.Name = s; return nil },
	)
	handleErr(err)

	pCity, err := param.New("City",
		func(s string) error { user.City = City(s); return nil },
		//Adding more options
		param.WithIsMandatory(true),
		param.WithDefault("Vancouver"),
		param.WithDesc("City where user lives"),
		param.WithExamples("Toronto", "Vancouver"),
		//When using command line arguments, uses `-Town=` to set this value
		param.WithFlag(param.WithFlagName("Town")),
		//When using environment variables, reads key `TOWN` to set this value
		param.WithEnvVar(param.WithEnvVarName("TOWN")),
		param.WithEnumValues("Toronto", "Vancouver", "Montreal"),
		// + other options like WithExclusive
	)
	handleErr(err)

	pAge, err := param.New("Age",
		func(s string) error {
			user.Age, err = strconv.Atoi(s)
			return err
		},
		//fetch data from another source. Could be a local file, a secret manager ...
		param.WithLoader(
			func(ctx context.Context) (string, error) {
				return "35", nil
			},
			//Check specific example to refresh regularly the value
		),
	)
	handleErr(err)

	c, err := config.New(config.WithParams(pName, pCity, pAge))
	handleErr(err)
	err = c.Init(
		context.Background(),
		//(For this example) Forcing what we receive in the command line. Default is os.Args[1:]
		config.WithInputArgs([]string{"-Name=Vincent"}),
	)
	handleErr(err)

	fmt.Printf("%+v", user)
	// Output:
	// {Name:Vincent City:Vancouver Age:35}
}

// Using the struct tags style (similar to the json package).
func Example_whenStructTagsStyle() {
	type City string
	user := struct {
		Name             string
		City             City `flag:"Town" envVar:"TOWN" mandatory:"true" desc:"City where user lives" examples:"Toronto;Vancouver" default:"Vancouver" enumValues:"Toronto;Vancouver;Montreal"`
		Age              int
		nonExportedField string `mandatory:"true"` //will be ignored (golang language: require start with CAPITAL LETTER to be exported)
	}{}

	//Read the tags on the struct field. And tries to match simple types in the automatic parse() function.
	c, err := config.New(config.WithParamsFromStructTag(&user, "prefix"))
	handleErr(err)
	err = c.Init(
		context.Background(),
		//(For this example) Forcing what we receive in the command line. Default is os.Args[1:]
		config.WithInputArgs([]string{"-prefixName=Vincent", "-prefixAge=35"}),
	)
	handleErr(err)

	fmt.Printf("%+v", user)
	// Output:
	// {Name:Vincent City:Vancouver Age:35 nonExportedField:}
}

// Using a command (like "git commit").
func Example_whenSubCommand() {
	type User struct {
		Name string
	}
	type UserAndAge struct {
		*User
		Age int
	}
	type UserAndCity struct {
		*User
		City string
	}

	user := User{}
	userAndAge := UserAndAge{
		User: &user,
	}
	userAndCity := UserAndCity{
		User: &user,
	}

	cUserAndAge, err := config.New(
		config.WithParamsFromStructTag(&userAndAge, ""),
		config.WithDescription("getting User.Name + City"),
		config.WithCallback(func() {
			//Triggers when this config (for this SubCommand) has been parsed.
		}),
	)
	handleErr(err)
	cUserAndCity, err := config.New(
		config.WithParamsFromStructTag(&userAndCity, ""),
	)
	handleErr(err)

	cUser, err := config.New(
		config.WithParamsFromStructTag(&user, ""),
		config.WithSubCommand("age", cUserAndAge),
		config.WithSubCommand("city", cUserAndCity),
		//don't error out if a City flag is provided and this config only declares the flag Name.
		config.WithIgnoreFlagProvidedNotDefined(true),
	)
	handleErr(err)

	err = cUser.Init(
		context.Background(),
		//(For this example) Forcing what we receive in the command line. Default is os.Args[1:]
		config.WithInputArgs([]string{"age", "-Name=Vincent", "-Age=35"}),
	)
	handleErr(err)

	//We could also call with command:"city" and use param "-City"
	//We could also call without command, using only the param "-Name"

	fmt.Printf("userAndAge: Name=%s Age=%d", userAndAge.Name, userAndAge.Age)
	// Output:
	// userAndAge: Name=Vincent Age=35
}

type user struct {
	Name string
}

func initConfigWithSync() func() user {
	var err error
	u := user{}
	loaderFakeValue := ""

	pName, err := param.New("Name",
		func(s string) error { u.Name = s; return nil },
		param.WithLoader(
			func(ctx context.Context) (string, error) {
				//simulating a constant change
				loaderFakeValue += "-"
				return loaderFakeValue, nil
			},
			// Check regularly in addition to startup
			// (For this example: super fast)
			param.WithSynchroFrequency(70*time.Millisecond),
		),
	)
	handleErr(err)

	lock := lock.New()
	c, err := config.New(
		config.WithParams(pName),
		config.WithLock(lock),
	)
	handleErr(err)
	err = c.Init(
		context.Background(),
		//(For this example) Forcing what we receive in the command line. Default is os.Args[1:]
		config.WithInputArgs([]string{}),
	)
	handleErr(err)

	return func() user {
		lock.Lock()
		defer lock.Unlock()
		return u
	}
}

func Example_whenLoaderSync() {
	userProxy := initConfigWithSync()

	//using a function with the lock to avoid race condition on value read.
	//This is only needed when using Loader with Synchro.
	//See also lock.LockWithContext(ctx)

	fmt.Printf("%+v\n", userProxy())
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("%+v\n", userProxy())
	// Output:
	// {Name:-}
	// {Name:--}
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

// When we need a part of the configuration to define another part. For example the name of the AWS secret manager key to use.
//
// Focus on `config.WithIgnoreFlagProvidedNotDefined(true)`
func Example_whenMultiSteps() {
	type applicationConfiguration struct {
		secretID    string
		secretValue string
	}

	appConfig := applicationConfiguration{}

	// first level, get secretID
	pSecretID, err := param.New(
		"secretID",
		func(s string) error { appConfig.secretID = s; return nil },
		param.WithDefault("SecretID_A"),
	)
	if err != nil {
		panic(err)
	}

	configManager, err := config.New(
		config.WithParams(pSecretID),
		//This is required in this case (a flag could be provided for an ulterior step and not be defined yet)
		config.WithIgnoreFlagProvidedNotDefined(true),
	)
	if err != nil {
		panic(err)
	}
	err = configManager.Init(
		context.Background(),
		//(For this example) Forcing what we receive in the command line. Default is os.Args[1:]
		config.WithInputArgs([]string{}),
	)
	if err != nil {
		panic(err)
	}

	//second step, using a map to simulate something dynamic
	m := map[string]string{"SecretID_A": "SecretValue_A"}

	pSecretValue, err := param.New("SecretValue",
		func(s string) error { appConfig.secretValue = s; return nil },
		param.WithLoader(
			func(ctx context.Context) (string, error) {
				return m[appConfig.secretID], nil
			},
		),
	)
	if err != nil {
		panic(err)
	}

	configManager, err = config.New(
		config.WithParams(pSecretValue),
	)
	if err != nil {
		panic(err)
	}
	err = configManager.Init(
		context.Background(),
		//(For this example) Forcing what we receive in the command line. Default is os.Args[1:]
		config.WithInputArgs([]string{}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("appConfig.secretID=%q appConfig.secretValue=%q\n", appConfig.secretID, appConfig.secretValue)
	// Output:
	// appConfig.secretID="SecretID_A" appConfig.secretValue="SecretValue_A"
}

// Example_whenMultipleSubCommand
func Example_whenMultipleSubCommand() {
	type User struct {
		Name string `mandatory:"true"`
	}
	type UserAndAge struct {
		*User
		Age int `mandatory:"true"`
	}
	type UserAndAgeAndCity struct {
		*UserAndAge
		City string `mandatory:"true"`
	}
	type UserAndAgeAndCityAndJob struct {
		*UserAndAgeAndCity
		Job string `mandatory:"true"`
	}

	user := User{}
	userAndAge := UserAndAge{
		User: &user,
	}
	userAndAgeAndCity := UserAndAgeAndCity{
		UserAndAge: &userAndAge,
	}
	userAndAgeAndCityAndJob := UserAndAgeAndCityAndJob{
		UserAndAgeAndCity: &userAndAgeAndCity,
	}

	cUserAndAgeAndCityAndJob, err := config.New(
		config.WithParamsFromStructTag(&userAndAgeAndCityAndJob, ""),
	)
	handleErr(err)

	cUserAndAgeAndCity, err := config.New(
		config.WithParamsFromStructTag(&userAndAgeAndCity, ""),
		config.WithSubCommand("job", cUserAndAgeAndCityAndJob),
		config.WithCallback(func() {
			//Triggers when this config (for this SubCommand) has been parsed.
		}),
	)
	handleErr(err)

	cUserAndAge, err := config.New(
		config.WithParamsFromStructTag(&userAndAge, ""),
		config.WithSubCommand("city", cUserAndAgeAndCity),
		//don't error out if a City|Job flag is provided and this config only declares the flag Name+Age.
		config.WithIgnoreFlagProvidedNotDefined(true),
		config.WithCallback(func() {
			//Triggers when this config (for this SubCommand) has been parsed.
			//Example blocking this sub command alone.
			panic(`fail use subcommand "age" alone. Either use no subcommand or use "age city" or use "age city job"`)
		}),
	)
	handleErr(err)

	cUser, err := config.New(
		config.WithParamsFromStructTag(&user, ""),
		config.WithSubCommand("age", cUserAndAge),
		//don't error out if a Age|City|Job flag is provided and this config only declares the flag Name+Age.
		config.WithIgnoreFlagProvidedNotDefined(true),
		config.WithCallback(func() {
			//Triggers when this config (for this SubCommand) has been parsed.
		}),
	)
	handleErr(err)

	err = cUser.Init(
		context.Background(),
		//(For this example) Forcing what we receive in the command line. Default is os.Args[1:]
		config.WithInputArgs([]string{"age", "city", "job", "-Job=dev", "-Name=Vincent", "-Age=35", "-City=Vancouver"}),
	)
	handleErr(err)

	fmt.Printf("Name:%q, Age:%d, City:%q, Job:%q",
		userAndAgeAndCityAndJob.Name, userAndAgeAndCityAndJob.Age, userAndAgeAndCityAndJob.City, userAndAgeAndCityAndJob.Job)

	err = cUser.Init(
		context.Background(),
		//(For this example) Forcing what we receive in the command line. Default is os.Args[1:]
		config.WithInputArgs([]string{"age", "city"}),
	)
	if err == nil {
		panic("expected error")
	}
	fmt.Println("\n\nAnd usage example with SubCommands =>")
	fmt.Println(err)

	// Output:
	// Name:"Vincent", Age:35, City:"Vancouver", Job:"dev"
	//
	// And usage example with SubCommands =>
	// ConfigWithUsageError: on SubCommands: [ age city], ConfigError for Param:"City": mandatory value
	// Usage:
	// 	Param: City
	// 		Mandatory value.
	// 		Command line flag: -City
	// 		Environment variable name: City
	// 		No custom loader defined.
}
