# "Configo"

Configuration manager in Golang. Populate your configuration struct from flags, env vars or external providers.
For example read from a secret manager and refresh the value every 12H. Or override with local value when running locally.

## Doc and examples

https://pkg.go.dev/github.com/vincentkerdraon/configo

## Competitors

To do quite the same purpose, see also:
- //FIXME

## TODO

- Auto completion in bash
- create param.NewBool calling param.New but with the parse func `func(b bool) error` + other common types
- reader JSON with sub keys and arrays
- parse() returns errors => show usage
- awssecretmanager BUG when same secret name used. (for example using different regions or accounts)
- add a logger
- deploy sub modules to https://pkg.go.dev/ (`GOPROXY=proxy.golang.org go list -m github.com/vincentkerdraon/configo/awssecretmanager@v0.1.0`)
