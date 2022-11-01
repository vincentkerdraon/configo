package secretrotation_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/vincentkerdraon/configo/secretrotation"
)

func TestSecretRotation(t *testing.T) {
	m := secretrotation.New()

	//Init not done
	_, err := m.RotatingSecret()
	if !errors.Is(err, secretrotation.MissingInitValuesError{}) {
		t.Fatal(err)
	}
	_, err = m.Current()
	if !errors.Is(err, secretrotation.MissingInitValuesError{}) {
		t.Fatal(err)
	}

	//Provide bad input: empty
	err = m.Set(secretrotation.NewRotatingSecret("", "my_secretB", "my_secretC"))
	if err == nil {
		t.Fatal()
	}

	//Provide good input
	err = m.Set(secretrotation.NewRotatingSecret("my_secretA", "my_secretB", "my_secretC"))
	if err != nil {
		t.Fatal(err)
	}

	secrets, err := m.RotatingSecret()
	if err != nil {
		t.Fatal(err)
	}

	if secrets.Previous != "my_secretA" || secrets.Current != "my_secretB" || secrets.Pending != "my_secretC" {
		t.Fatalf("%+v", secrets)
	}

	secret, err := m.Current()
	if err != nil {
		t.Fatal(err)
	}
	if secret != "my_secretB" {
		t.Fatalf("%s", secret)
	}

	//Rotate
	err = m.Set(secretrotation.NewRotatingSecret("my_secretB", "my_secretC", "my_secretD"))
	if err != nil {
		t.Fatal(err)
	}
	secret, err = m.Current()
	if err != nil {
		t.Fatal(err)
	}
	if secret != "my_secretC" {
		t.Fatalf("%s", secret)
	}

	//Rotate
	err = m.Set(secretrotation.NewRotatingSecret("my_secretC", "my_secretD", "my_secretE"))
	if err != nil {
		t.Fatal(err)
	}
	secret, err = m.Current()
	if err != nil {
		t.Fatal(err)
	}
	if secret != "my_secretD" {
		t.Fatalf("%s", secret)
	}

	//Allowed?
	if m.Allowed("wrong") {
		t.Fatal()
	}
	if !m.Allowed("my_secretE") {
		t.Fatal()
	}
	if !m.AllowedNonConstant("my_secretE") {
		t.Fatal()
	}
}

// goos: linux
// goarch: amd64
// pkg: github.com/vincentkerdraon/configo/secretrotation
// cpu: AMD Ryzen 9 3950X 16-Core Processor
// 20220912
// BenchmarkSecretRotation-32    	 1000000	      1834 ns/op	     417 B/op	       3 allocs/op
// BenchmarkSecretRotation-32    	 1000000	      1630 ns/op	     401 B/op	       3 allocs/op
// 20221004
// BenchmarkSecretRotation-32    	 1214138	       935.9 ns/op	     191 B/op	       1 allocs/op
// BenchmarkSecretRotation-32    	 1000000	      1013 ns/op	     255 B/op	       1 allocs/op
func BenchmarkSecretRotation(b *testing.B) {
	//Not really a benchmarch, used to detect races.

	m := secretrotation.New()

	//start with non empty value
	err := m.Set(secretrotation.NewRotatingSecret("my_secretA", "my_secretB", "my_secretC"))
	if err != nil {
		b.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(b.N)
	for n := 0; n < b.N; n++ {
		go func() {
			defer wg.Done()
			err = m.Set(secretrotation.NewRotatingSecret("my_secretA", "my_secretB", "my_secretC"))
			if err != nil {
				b.Error(err)
			}

			secret, err := m.Current()
			if err != nil {
				b.Error(err)
			}

			if secret != "my_secretB" {
				b.Errorf("%s", secret)
			}
		}()
	}
	wg.Wait()
}
