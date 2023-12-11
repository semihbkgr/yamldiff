package yaml

type Type int

const (
	UnknownType Type = iota
)

// Token represents a token in YAML.
type Token struct {
	Type   Type
	Value  string
	Line   int
	Indent int
}
