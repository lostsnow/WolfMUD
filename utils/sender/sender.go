package sender

type Interface interface {
	Send(format string, any ...interface{})
	SendWithoutPrompt(format string, any ...interface{})
}
