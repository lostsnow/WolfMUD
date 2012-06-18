package startingLocation

import (
	"math/rand"
	"wolfmud.org/entities/location"
)

var startingLocations []*StartingLocation

func GetStart() *StartingLocation {
	return startingLocations[rand.Intn(len(startingLocations))]
}

type StartingLocation struct {
	*location.Location
}

func New(name string, aliases []string, description string) *StartingLocation {
	l := &StartingLocation{
		Location: location.New(name, aliases, description),
	}
	startingLocations = append(startingLocations, l)
	return l
}
