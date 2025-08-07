package tool

type Tool interface {
	Name() string
	Description() string
	Params() map[string]any
	Call(input string) (string, error)
}
