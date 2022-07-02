// Copyright 2022 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"code.wolfmud.org/WolfMUD.git/text"
)

var roundDuration = (3 * time.Second).Nanoseconds()

func createCorpse(t *Thing) *Thing {
	c := NewThing()
	c.As[Name] = "the corpse of " + t.As[Name]
	c.As[UName] = "The corpse of " + t.As[Name]
	c.As[TheName] = "the corpse of " + t.As[Name]
	c.As[UTheName] = "The corpse of " + t.As[Name]
	c.As[Description] = t.As[Description]
	c.Any[Alias] = append(c.Any[Alias], t.Any[Alias]...)
	c.Any[Qualifier] = append(c.Any[Qualifier], t.Any[Qualifier]...)
	c.Ref[Where] = t.Ref[Where]
	c.Ref[Where].In[c.As[UID]] = t
	c.Int[CleanupAfter] = time.Duration(60 * time.Second).Nanoseconds()
	c.As[OnCleanup] = c.As[UTheName] + " turns to dust."

	// Replace original UID alias with "CORPSE" (new UID was added by NewThing)
	for x, alias := range c.Any[Alias] {
		if alias == t.As[UID] {
			c.Any[Alias][x] = "CORPSE"
			break
		}
	}

	return c
}

func (s *state) Attack() {

	if len(s.word) == 0 {
		s.Msg(s.actor, text.Info, "You go to attack... someone?")
		return
	}

	if len(s.actor.Any[Opponents]) > 0 {
		s.Msg(s.actor, text.Bad, "You are already fighting!")
		return
	}

	where := s.actor.Ref[Where]
	if len(where.Who) >= cfg.crowdSize {
		s.Msg(s.actor, text.Bad, "It's too crowded to start a fight here!")
		return
	}

	uids := Match(s.word, where)
	uid := uids[0]
	what := where.Who[uid]
	if what == nil {
		what = where.In[uid]
	}

	switch {
	case what == nil:
		s.Msg(s.actor, text.Bad, "You see no '", uid, "' here to attack.")
	case s.actor == what:
		s.Msg(s.actor, text.Good, "You give yourself a slap. Awake now?")
		s.Msg(where, text.Info, s.actor.As[UName], " slaps themself.")
	case what.Is&(Player|NPC) == 0:
		s.Msg(s.actor, text.Bad, "You cannot fight ", what.As[TheName], ".")
		s.Msg(where, text.Info, s.actor.As[UName], " tries to attack ", what.As[Name], ".")
	case where.As[VetoCombat] != "":
		s.Msg(s.actor, text.Bad, where.As[VetoCombat])
	default:
		what.Any[Opponents] = append(what.Any[Opponents], s.actor.As[UID])
		what.Suspend(Action)

		s.actor.Any[Opponents] = append(s.actor.Any[Opponents], what.As[UID])
		s.actor.Ref[Opponent] = what
		s.actor.Int[CombatAfter] = roundDuration
		s.actor.Schedule(Combat)

		s.Msg(s.actor, text.Good, "You attack ", what.As[TheName], "!")
		s.Msg(what, text.Bad, s.actor.As[TheName], " attacks you!")
		s.Msg(where, text.Info, s.actor.As[UTheName], " attacks ", what.As[TheName], "!")
	}
}

func (s *state) Combat() {

	what := s.actor.Ref[Opponent]
	where := s.actor.Ref[Where]

	if what == nil || where != what.Ref[Where] {
		s.stopCombat(s.actor, nil)
		s.Msg(s.actor, text.Info, "\nYou stop fighting, your opponent disappeared...")
		return
	}

	attacker, defender := s.actor, what
	if roll := rand.Float64(); roll > s.hitChance(attacker, defender) {
		attacker, defender = defender, attacker
	}

	damage := damageFixed(attacker) + rand.Int63n(damageRandom(attacker)+1)
	damageText := fmt.Sprintf(" doing %d damage.", damage)
	defender.Int[HealthCurrent] -= damage

	s.MsgAppend(attacker, text.Good, "You hit ", defender.As[TheName], damageText)
	s.MsgAppend(defender, text.Bad, attacker.As[UTheName], " hits you", damageText)
	s.Msg(where, text.Info, attacker.As[UTheName], " hits ", defender.As[Name], ".")

	// defender not killed, do health bookkeeping and go another round
	if defender.Int[HealthCurrent] > 0 {
		s.StatusUpdate(defender)
		if defender.Event[Health] == nil {
			defender.Schedule(Health)
		}
		s.actor.Int[CombatAfter] = roundDuration
		s.actor.Schedule(Combat)
		return
	}
	s.Msg(attacker, "You kill ", defender.As[Name], "!")
	s.Msg(defender, attacker.As[UTheName], " kills you!")
	s.Msg(where, attacker.As[UTheName], " kills ", defender.As[Name], "!")

	// Stop everyone attacking defender and notify them, as they receive a
	// specific message they won't get the message to the location.
	for _, uid := range defender.Any[Opponents] {
		who := where.Who[uid]
		if who != nil && who != attacker {
			s.Msg(who, text.Info, attacker.As[UTheName], " kills ", defender.As[Name], "!")
			s.Msg(who, text.Info, "You stop fighting ", defender.As[Name], ".")
		}
		if who == nil {
			who = where.In[uid]
		}
		s.stopCombat(who, defender)
	}
	s.stopCombat(defender, nil)

	// Create and place corpse
	c := createCorpse(defender)
	where.In[c.As[UID]] = c
	c.Schedule(Cleanup)

	// Remove defender from location
	delete(where.Who, defender.As[UID])

	// If not a player junk for a reset
	if defender.Is&Player == 0 {
		defender.Junk()
		return
	}

	// Place player back into the world
	start := WorldStart[rand.Intn(len(WorldStart))]
	defender.Int[HealthCurrent] = 1
	defender.Ref[Where] = start
	start.Who[defender.As[UID]] = defender

	if s.actor == defender {
		s.subparse("$POOF")
	} else {
		s.subparseFor(defender, "$POOF")
	}
}

func (s *state) stopCombat(who, what *Thing) {
	if who == nil {
		return
	}
	if what == nil {
		who.Cancel(Combat)
		who.Schedule(Action)
		delete(who.Ref, Opponent)
		delete(who.Any, Opponents)
	} else {
		who.Any[Opponents], _ = remainder(who.Any[Opponents], []string{what.As[UID]})
		if len(who.Any[Opponents]) == 0 {
			who.Schedule(Action)
			who.Cancel(Combat)
			delete(who.Ref, Opponent)
			delete(who.Any, Opponents)
		}
	}
}

var (
	chanceMin, _ = new(big.Float).SetString("0.000001")
	chanceMax, _ = new(big.Float).SetString("0.999999")
)

// hitChance returns the percentage, as a decimal float, chance of an attacker
// hitting a defender. Chance range is 0.000001-0.999999 or 0.0001%-99.9999%.
func (s *state) hitChance(Attacker, Defender *Thing) float64 {

	Aatk, Adef := attack(Attacker), defense(Attacker)
	Datk, Ddef := attack(Defender), defense(Defender)

	Aatk_Aatk := new(big.Float).SetInt64(Aatk * Aatk)
	Aatk_Ddef := new(big.Float).SetInt64(Aatk + Ddef)
	Datk_Datk := new(big.Float).SetInt64(Datk * Datk)
	Datk_Adef := new(big.Float).SetInt64(Datk + Adef)

	// score = attack * attack / ( attack + opponent defense )
	Ascore := new(big.Float).Quo(Aatk_Aatk, Aatk_Ddef)
	Dscore := new(big.Float).Quo(Datk_Datk, Datk_Adef)

	Ascore_Dscore := new(big.Float).Add(Ascore, Dscore)

	// chance% = score / ( score + opponent score )
	Achance := new(big.Float).Quo(Ascore, Ascore_Dscore)
	Dchance := new(big.Float).Quo(Dscore, Ascore_Dscore)

	// Allow for 1000000:1 chances to miss, or hit...
	if Achance.Cmp(chanceMax) == 1 {
		Achance.Copy(chanceMax)
	}
	if Achance.Cmp(chanceMin) == -1 {
		Achance.Copy(chanceMin)
	}

	_ = Dchance // Used for debugging only, prevent compiler error

	chance, _ := Achance.Float64()
	return chance
}

// attack returns the total attack value for an actor. Items can provide
// positive or negative contributions. Minimum attack is 1.0.
func attack(actor *Thing) int64 {

	// Note: integer division will truncate towards zero, we round up
	d := damageRandom(actor)/2 + damageRandom(actor)%2
	a := damageFixed(actor) + d + actor.Int[HealthCurrent]
	if a < 1 {
		a = 1
	}
	return a
}

// defense returns the total defense value for an actor. Items can provide
// positive or negative contributions. Minimum defense is 1.0.
func defense(actor *Thing) int64 {
	d := actor.Int[Armour] + actor.Int[HealthCurrent]
	for _, item := range actor.In {
		if item.Is&(Wearing|Wielding) != 0 {
			d += item.Int[Armour]
		}
	}
	if d < 1 {
		d = 1
	}
	return d
}

// damageFixed returns the total amount of fixed damage an actor can cause.
// This includes natural and wielded/worn item damage. Items can provide
// positive or negative contributions. Minimum damage is 1.
func damageFixed(actor *Thing) int64 {
	df := actor.Int[DamageFixed]
	for _, item := range actor.In {
		if item.Is&(Wearing|Wielding) != 0 {
			df += item.Int[DamageFixed]
		}
	}
	if df < 1 {
		df = 1
	}
	return df
}

// damageRandom returns the total amount of random damage an actor can cause.
// This includes naturnal and wielded/worn item damage. Items can provide
// positive or negative contributions. Minimum damage is 0.
func damageRandom(actor *Thing) int64 {
	dr := actor.Int[DamageRandom]
	for _, item := range actor.In {
		if item.Is&(Wearing|Wielding) != 0 {
			dr += item.Int[DamageRandom]
		}
	}
	if dr < 0 {
		dr = 0
	}
	return dr
}
