# "Configo"

Configuration manager in Golang. Populate your configuration struct from flags, env vars or external providers.
For example read from a secret manager and refresh the value every 12H. Or override with local value when running locally.

## Doc and examples

https://pkg.go.dev/github.com/vincentkerdraon/configo

## Competitors

To do quite the same purpose, see also:
- https://micro.dev/blog/2018/07/04/go-config.html
- https://github.com/spf13/viper + https://github.com/spf13/cobra

## TODO

- Auto completion in bash
- create param.NewBool calling param.New but with the parse func `func(b bool) error` + other common types
- reader JSON with sub keys and arrays
- awssecretmanager BUG when same secret name used. (for example using different regions or accounts)
- add a logger
- improve Competitors list
- awssecretmanager lambda and doc
- WithCallback should be `func() error`
