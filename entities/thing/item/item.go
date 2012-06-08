package item

import (
	"bytes"
	"strconv"
	"wolfmud.org/entities/location"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/command"
	"wolfmud.org/utils/inventory"
)

type Interface interface {
}

// ounces is used as the standard weight for items. For a more modern game
// setting the weight can easily be called something else. In which case only
// the String method needs updating and the Item.weight definition.
type ounces int

// String displays ounces as pounds and ounces. Ounces are only displayed for
// light weights. If the weight is 2 pounds or more then the ounces are not
// displayed but the pounds are rounded up if there are over 8 ounces. For
// example:
//
//	2 ounces is displayed as "2oz"
//	18 ounces displays as "1lb and 2oz"
//	88 ounces displays as "5lb" and not "5lb and 8oz"
//	89 ounces displays as "6lb" and not "5lb and 9oz"
//
func (o ounces) String() string {
	b := new(bytes.Buffer)

	o_int := int(o)

	oz := o_int % 16
	lb := (o_int - oz) / 16

	if lb >= 2 && oz > 8 {
		lb++
	}

	if lb != 0 {
		b.WriteString(strconv.Itoa(lb))
		b.WriteString("lb")
	}
	if oz != 0 && lb < 2 {
		if b.Len() != 0 {
			b.WriteString(" and ")
		}
		b.WriteString(strconv.Itoa(oz))
		b.WriteString("oz")
	}

	return b.String()
}

type Item struct {
	*thing.Thing
	weight ounces
}

func New(name string, aliases []string, description string, weight ounces) *Item {
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
