package item

import (
	"wolfmud.org/entities/location"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/command"
	"wolfmud.org/utils/inventory"
	"wolfmud.org/utils/units"
)

type Interface interface {
}

type Item struct {
	*thing.Thing
	weight units.Weight
}

func New(name string, aliases []string, description string, weight units.Weight) *Item {
	return &Item{
		Thing:  thing.New(name, aliases, description),
		weight: weight,
	}
}

func (i *Item) Process(cmd *command.Command) (handled bool) {

	if cmd.Target == nil || !i.IsAlias(*cmd.Target) {
		return
	}

	switch cmd.Verb {
	case "DROP":
		handled = i.drop(cmd)
	case "WEIGH":
		handled = i.weigh(cmd)
	case "EXAMINE", "EXAM":
		handled = i.examine(cmd)
	case "GET":
		handled = i.get(cmd)
	case "JUNK":
		handled = i.junk(cmd)
	}

	return
}

func (i *Item) drop(cmd *command.Command) (handled bool) {
	if m, ok := cmd.Issuer.(location.Locateable); ok {
		if inv, ok := cmd.Issuer.(inventory.Interface); ok {
			if inv.Contains(i) {
				inv.Remove(i)
				cmd.Respond("You drop %s.", i.Name())

				l := m.Locate()
				l.Add(i)
				l.Broadcast([]thing.Interface{cmd.Issuer}, "You see %s drop %s.", cmd.Issuer.Name(), i.Name())

				handled = true
			}
		}
	}
	return
}

func (i *Item) weigh(cmd *command.Command) (handled bool) {

	cmd.Respond(
		"You estimate the weight of %s to be about %s.",
		i.Name(),
		i.weight,
	)

	return true
}

func (i *Item) examine(cmd *command.Command) (handled bool) {
	cmd.Respond("You examine %s. %s", i.Name(), i.Description())
	return true
}

func (i *Item) get(cmd *command.Command) (handled bool) {
	if m, ok := cmd.Issuer.(location.Locateable); ok {
		if inv, ok := cmd.Issuer.(inventory.Interface); ok {
			if l := m.Locate(); l.Contains(i) {
				l.Remove(i)
				l.Broadcast([]thing.Interface{cmd.Issuer}, "You see %s pick up %s.", cmd.Issuer.Name(), i.Name())

				inv.Add(i)
				cmd.Respond("You pickup %s.", i.Name())

				handled = true
			}
		}
	}
	return
}

func (i *Item) junk(cmd *command.Command) (handled bool) {
	cmd.Respond("Junk not implemented yet.", i.Name())
	return
}
