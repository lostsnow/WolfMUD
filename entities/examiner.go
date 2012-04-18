package entities

type Examiner interface {
	examine(what Thing, args []string) (handled bool)
}
