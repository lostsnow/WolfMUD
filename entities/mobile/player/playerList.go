package player

import (
	"fmt"
	"log"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/text"
)

type playerList struct {
	players []*Player
	mutex   chan bool
}

var (
	PlayerList playerList
)

func init() {
	PlayerList.mutex = make(chan bool, 1)
}

func (l *playerList) lock() {
	l.mutex <- true
}

func (l *playerList) unlock() {
	<-l.mutex
}

func (l *playerList) Add(player *Player) {
	l.lock()
	defer l.unlock()
	l.players = append(l.players, player)
}

func (l *playerList) Remove(player *Player) {
	l.lock()
	defer l.unlock()
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
	l.lock()
	defer l.unlock()
	return len(l.players)
}

func (l *playerList) List(omit ...thing.Interface) (list []*Player) {
	l.lock()
	defer l.unlock()

  return l.nonLockingList(omit...)
}

func (l *playerList) Broadcast(omit []thing.Interface, format string, any ...interface{}) {
	l.lock()
	defer l.unlock()

	msg := text.Colorize(fmt.Sprintf("\n"+format, any...))

	for _, p := range l.nonLockingList(omit...) {
		p.Respond(msg)
	}
}

func (l *playerList) nonLockingList(omit ...thing.Interface) (list []*Player) {

OMIT:
	for _, player := range l.players {
		for i, o := range omit {
			if player.IsAlso(o) {
				omit = append(omit[0:i], omit[i+1:]...)
				continue OMIT
			}
		}
		list = append(list, player)
	}

	return
}
