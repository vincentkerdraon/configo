# "Configo"

Configuration manager in Golang. Populate your configuration struct from flags, env vars or external providers.
For example read from a secret manager and refresh the value every 12H. Or override with local value when running locally.

## Doc and examples

See godoc

## Competitors

To do quite the same purpose, see also:
- //FIXME

## TODO

- Auto completion in bash
- Helper for the aws instance tag reading
- Allow detection when using a command (a param is not necessary needed, "git commit")
- Improve usage(). Especially for sub commands.
- create param.NewBool calling param.New but with the parse func `func(b bool) error` + other common types
- reader JSON with sub keys and arrays
- parse() returns errors => show usage
- awssecretmanager BUG when same secret name used. (for example using different regions or accounts)
- add a logger
