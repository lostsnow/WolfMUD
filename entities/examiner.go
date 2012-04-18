package entities

type Examiner interface {
	examine(c Cmd) (handled bool)
}
