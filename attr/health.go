// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"log"
	"time"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
)

// Register marshaler for Health attribute.
func init() {
	internal.AddMarshaler((*Health)(nil), "health")
}

// Health implements an attribute representing either the state of health of a
// Thing or health bonuses/penalities to be applied to a Thing. If the Health
// attribute is on a Player or Mobile then the values are absolute and
// represent current and maximum health levels of the Player or Mobile. If the
// Health attribute is on anything else the values are relative modifiers
// applied to a Player or Mobile when the Thing is used, worn, wielded, eaten,
// drunk or otherwise applied.
type Health struct {
	Attribute
	current   int
	maximum   int
	regens    int
	frequency int64
	update    int64
}

// Some interfaces we want to make sure we implement
var (
	_ has.Health = &Health{}
)

// NewHealth returns a new Health attribute. If the Health attribute is added
// to a Player the current, maximum, regens and frequency (in seconds) values
// are absolute values representing the base values of the Player. Otherwise
// the values are relative and modify a Player's base values when applicable.
//
// The frequency is how often (in seconds) health regenerates. So a value of 10
// is every 10 seconds while a value of 90 is every 1 minute 30 seconds.
// Smaller values increase the number of updates while larger values decrease
// the number of updates - for a given period of time.
//
// For example a ring of healing may have frequency=-5 and regens=+2 to
// increase the frequency the Player regenerates health and increase the amount
// they regenerate - so the Player regenerates more health quicker - but the
// effects only apply when the ring is being worn by the Player.
func NewHealth(current, maximum, regens int, frequency int64) *Health {
	return &Health{Attribute{}, current, maximum, regens, frequency, 0}
}

// FindHealth searches the attributes of the specified Thing for attributes that
// implement has.Health returning the first match it finds or a *Health typed nil
// otherwise.
func FindHealth(t has.Thing) has.Health {
	return t.FindAttr((*Health)(nil)).(has.Health)
}

// Is returns true if passed attribute implements Health else false.
func (*Health) Is(a has.Attribute) bool {
	_, ok := a.(has.Health)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (n *Health) Found() bool {
	return n != nil
}

// Unmarshal is used to turn the passed data into a new Health attribute.
func (*Health) Unmarshal(data []byte) has.Attribute {
	h := NewHealth(0, 0, 0, 0)
	for field, data := range decode.PairList(data) {
		data := []byte(data)
		switch field {
		case "MAX", "MAXIMUM":
			h.maximum = decode.Integer(data)
		case "CUR", "CURRENT":
			h.current = decode.Integer(data)
		case "FREQ", "FREQUENCY":
			h.frequency = int64(decode.Duration(data).Seconds())
		case "REGENS", "REGENERATES":
			h.regens = decode.Integer(data)
		default:
			log.Printf("Health.unmarshal unknown attribute: %q: %q", field, data)
		}
	}
	return h
}

// Marshal returns a tag and []byte that represents the receiver.
func (h *Health) Marshal() (tag string, data []byte) {
	return "health", encode.PairList(
		map[string]string{
			"current":     string(encode.Integer(h.current)),
			"maximum":     string(encode.Integer(h.maximum)),
			"frequency":   string(encode.Duration(time.Duration(h.frequency) * time.Second)),
			"regenerates": string(encode.Integer(h.regens)),
		},
		'→',
	)
}

func (h *Health) Dump() []string {

	var tmpl string

	// Values are absolute if Health attribute is on a player otherwise relative
	// TODO(diddymus): And mobiles...
	absolute := FindPlayer(h.Parent()).Found()

	if absolute {
		tmpl = "%p %[1]T current: %d, maximum: %d, regens %d, frequency: %d"
	} else {
		tmpl = "%p %[1]T current: %+d, maximum: %+d, regens %+d, frequency: %+d"
	}

	return []string{DumpFmt(tmpl, h, h.current, h.maximum, h.regens, h.frequency)}
}

// State returns the current and maximum health points.
func (h *Health) State() (current, maximum int) {
	if h == nil {
		return 0, 0
	}
	h.regen()
	return h.current, h.maximum
}

// Adjust increases or decreses the current health points by the given amount.
// The new value will be a minimum of 0 and capped at the health maximum.
func (h *Health) Adjust(amount int) {
	if h == nil {
		return
	}

	h.regen()
	h.current += amount

	switch {
	case h.current < 0:
		h.current = 0
	case h.current > h.maximum:
		h.current = h.maximum
	}
}

// regen is responsible for regenerating current health points periodically.
// Health points regenerate at a rate of Health.regens per Health.frequency.
// Instead of regenerating on a timer regen calculates the number of updates
// that would have occurred since the last update and applies the applicable
// number of Health.regens to Health.current. As a result regens should be
// called before Health.current is read or updated.
func (h *Health) regen() {
	if h == nil {
		return
	}

	now := time.Now().Unix()

	// If not time for an update yet make a quick exit
	if now < h.update {
		return
	}

	// If next update not set yet just record next expected update and make a
	// quick exit
	if h.update == 0 {
		h.update = now + h.frequency
		return
	}

	// Calculate the difference between now and when we expected the next update
	// to occur. Also calculate fraction of latest update period passed.
	diff := now - h.update
	frac := diff % h.frequency

	// Set next expected update minus the fractional period that has already
	// passed
	h.update = now + h.frequency - frac

	// If health is already at its maximum exit now having just recorded when to
	// expect the next update
	if h.current == h.maximum {
		return
	}

	// Calculate number of complete update periods that have passed since last
	// update and adjust the current health points
	periods := (diff / h.frequency) + 1
	h.current += int(periods) * h.regens

	// Make sure current health points are capped at the maximum
	if h.current > h.maximum {
		h.current = h.maximum
	}

	return

}

// Copy returns a copy of the Health receiver.
func (h *Health) Copy() has.Attribute {
	if h == nil {
		return (*Health)(nil)
	}
	return NewHealth(h.current, h.maximum, h.regens, h.frequency)
}
