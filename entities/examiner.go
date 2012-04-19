package entities

type Examiner interface {
	examine(cmd Command) (handled bool)
}
