package subcommand

type (
	SubCommand string
)

func (n SubCommand) String() string {
	return string(n)
}
