package entities

type Looker interface {
	look(what Thing, args []string) (handled bool)
}
