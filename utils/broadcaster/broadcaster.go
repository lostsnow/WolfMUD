package broadcaster

import (
	"wolfmud.org/entities/thing"
)

type Interface interface {
	Broadcast(omit []thing.Interface, format string, any ...interface{})
	AddThing(thing thing.Interface)
}
