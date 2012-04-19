package entities

type Looker interface {
	look(cmd Command) (handled bool)
}
