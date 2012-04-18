package entities

type Looker interface {
	look(c Cmd) (handled bool)
}
