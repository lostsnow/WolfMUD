package entities

type Commander interface {
	Command(what Thing, cmd string, args []string) (handled bool)
}

type Cmd struct {
	what Thing
	cmd  string
	args []string
}
