package responder


type Interface interface {
	Respond(format string, any ...interface{})
}
