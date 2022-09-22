package paramname

type (
	ParamName string
)

func (n ParamName) String() string {
	return string(n)
}
