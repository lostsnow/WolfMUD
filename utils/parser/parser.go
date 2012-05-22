package parser

type Interface interface {
	Parse(input string)
	Name() (name string)
	Destroy()
}
