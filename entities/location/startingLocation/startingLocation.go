package startingLocation

import (
	"wolfmud.org/entities/location"
)

type StartingLocation struct {
	*location.Location
}

func New(name string, aliases []string, description string) *StartingLocation {
	return &StartingLocation{
		Location: location.New(name, aliases, description),
	}
}
