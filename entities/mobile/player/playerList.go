package player

import (
	"fmt"
	"log"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/responder"
)

type playerList struct {
	players []*Player
	lock chan bool
}

var (
	PlayerList playerList
)

func init() {
	PlayerList.lock = make(chan bool, 1)
}

func (l *playerList) Add(player *Player) {
	l.lock <- true
	defer func(){<-l.lock}()
	l.players = append(l.players, player)
}

func (l *playerList) Remove(player *Player) {
	l.lock <- true
	defer func(){<-l.lock}()
	found := false
	for index, p := range l.players {
		if player.IsAlso(p) {
			l.players = append(l.players[:index], l.players[index+1:]...)
			found = true
			break
		}
	}
	if !found {
		log.Printf("EEP!!! %s Not found to remove", player.Name())
	}
}

func (l *playerList) Length() int {
	l.lock <- true
	defer func(){<-l.lock}()
	return len(l.players)
}

func (l *playerList) List(omit ...thing.Interface) (list []*Player) {
	l.lock <- true
	defer func(){<-l.lock}()

OMIT:
	for _, player := range l.players {
		for _, o := range omit {
			if player.IsAlso(o) {
				continue OMIT
			}
		}
		list = append(list, player)
	}

	return
}

func (l *playerList) Broadcast(omit []thing.Interface, format string, any ...interface{}) {
	msg := fmt.Sprintf("\n"+format, any...)

	for _, t := range l.List(omit...) {
		t.Respond(msg)
	}
}
