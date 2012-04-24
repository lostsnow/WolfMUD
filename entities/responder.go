package entities

type Responder interface {
	Respond(format string, any ...interface{})
	RespondGroup(ommit []Thing, format string, any ...interface{})
}

type responder struct {
}

func (r *responder) Respond(format string, any ...interface{}) {
}

func (r *responder) RespondGroup(ommit []Thing, format string, any ...interface{}) {
}
