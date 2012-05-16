package player

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
	for index, p := range l.players {
		if p == player {
			l.players = append(l.players[:index], l.players[index+1:]...)
			break
		}
	}
}

func (l *playerList) Length() int {
	return len(l.players)
}

func (l *playerList) List() (pl []*Player) {
	pl = make([]*Player, len(l.players))
	copy(pl, l.players)
	return
}
