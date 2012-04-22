package entities

import (
	"fmt"
)

type Responder interface {
	Respond(format string, any ...interface{})
}

type responder struct {
}

func (r *responder) Respond(format string, any ...interface{}) {
	fmt.Printf(format, any...)
}
