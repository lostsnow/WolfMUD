package player

import (
	"log"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/responder"
)

type playerList struct {
	players []*Player
}

var (
	PlayerList playerList
)

func (l *playerList) Add(player *Player) {
	l.players = append(l.players, player)
}

func (l *playerList) Remove(player *Player) {
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
	return len(l.players)
}

func (l *playerList) List(ommit ...thing.Interface) (list []*Player) {

OMMIT:
	for _, player := range l.players {
		for _, o := range ommit {
			if player.IsAlso(o) {
				continue OMMIT
			}
		}
		list = append(list, player)
	}

	return
}
