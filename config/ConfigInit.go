package config

type (
	Init struct {
		//InputArgs to read the flag input
		//
		// default: os.Args[1:] (with Args[0] being the name of the program)
		InputArgs []string //TODO subcommand mandatory, subcommand used?
	}

	configInitOptions func(r *Init) error
)

// WithInputArgs to define the flag input
//
// default: os.Args[1:] (with Args[0] being the name of the program)
func WithInputArgs(args []string) configInitOptions {
	return func(ci *Init) error {
		ci.InputArgs = args
		return nil
	}
}
