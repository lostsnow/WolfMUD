// Copyright 2022 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"code.wolfmud.org/WolfMUD.git/core"
	"code.wolfmud.org/WolfMUD.git/recordjar"
)

func Usage() {
	o := flag.CommandLine.Output()
	fmt.Fprintf(o, "Usage of %s:\n", filepath.Base(os.Args[0]))
	fmt.Fprintln(o, "\n  msgfmt -a [N|M|F|I] -d [N|M|F|I] -f [FILE] | TEXT...\n")
	flag.PrintDefaults()

	fmt.Fprintln(o, `
msgfmt is a utility for checking the syntax, formatting and replacements used
for messages. The attacker and defender used can be specified using the -a and
-d flags. For the available attackers and defenders, and their genders, that
can be used see below. msgfmt will then display the messages as seen by the
attacker, defender and any observers. For example:


  > msgfmt -a m -d i [%A] hit[/s] [%d] wounding [%d.them].
      Bob, attacker: You hit the imp wounding it.
   An imp, defender: Bob hits you wounding you.
           observer: Bob hits the imp wounding it.


  > msgfmt -a f -d n [%A] hit[/s] [%d] wounding [%d.them].
    Alice, attacker: You hit Charlie wounding them.
  Charlie, defender: Alice hits you wounding you.
           observer: Alice hits Charlie wounding them.


The message texts may be processed in bulk by redirecting the standard input.
The input should consist of one message per line. When more than one message
is processed the raw message and line number will be included in the output
unless supressed with the -R flag:


  > cat -n messages.txt
     1  [%A] hit[/s] [%d].
     2  [%A] hit[/s] [%d] wounding [%d.them].
  > msgfmt -a f -d m < messages.txt

      raw message 1: [%A] hit[/s] [%d].
    Alice, attacker: You hit Bob.
      Bob, defender: Alice hits you.
           observer: Alice hits Bob.

      raw message 2: [%A] hit[/s] [%d] wounding [%d.them].
    Alice, attacker: You hit Bob wounding him.
      Bob, defender: Alice hits you wounding you.
           observer: Alice hits Bob wounding him.

As an alternative to redirection a file can be specified via the -f option.

Messages contain substitution blocks of the form:

  [attacker/defender/observer]

Where attacker is only shown to attackers, defender to defenders and observer
to observers. If observer is omitted the defender replacement will be used. If
defender is omitted the attacker replacement will be used. This can be
overridden by providing an empty replacement.

The replacement text may contain %A or %D which will be replaced with the name
of the attacker and defender respectively. If specified uppercased the initial
letter of the name will be uppercased. If specified lowercased the replacement
will be used "as is". If the message is for the attacker %A will be replaced
with "you". If the message is for the defender %D will be replaced with "you".
In each instance "you" will be capitalised based on the case of %A and %D.

Available attackers and defenders are:

                              N(eutral) - Charlie
                              M(ale)    - Bob
                              F(emale)  - Alice
                              I(t)      - an imp

A gender neutral pronoun may follow %A or %D separated by a single period '.'
character. This will be replaced with a pronoun appropriate for the attacker's
or defender's gender. The pronoun will be cased according to the case of %A or
%D used.

Available neutral pronouns and their substitutions are:

                     Gender   Pronouns
                     -------  ---------------------------
                     NEUTRAL  they  them  their  themself
                     MALE     he    him   his    himself
                     FEMALE   she   her   her    herself
                     IT       it    it    its    itself


Examples:


  > msgfmt -a m -d i [%A] hit[/s] [%d] wounding [%d.them].
        Bob, attacker: You hit the imp wounding it.
     An imp, defender: Bob hits you wounding you.
             observer: Bob hits the imp wounding it.

  > msgfmt -a m -d i [%A] slash[/es] [%d] with [your/%a.their] dagger wounding [%d.them].
    Alice, attacker: You slash Charlie with your dagger wounding them.
  Charlie, defender: Alice slashes you with her dagger wounding you.
           observer: Alice slashes Charlie with her dagger wounding them.

  > msgfmt -a f -d m [%D] stumble[s//s], giving [%a] a chance to hit [%d.them].
    Alice, attacker: Bob stumbles, giving you a chance to hit him.
      Bob, defender: You stumble, giving Alice a chance to hit you.
           observer: Bob stumbles, giving Alice a chance to hit him.

  > msgfmt -a f -d m [%D] lash[es//es] out at [%a], but misjudge[s//s] \
      [%d.their][/r/] attack and injure[s//s] [%d.themself/yourself/%d.themself].
    Alice, attacker: Bob lashes out at you, but misjudges his attack and injures himself.
      Bob, defender: You lash out at Alice, but misjudge your attack and injure yourself.
           observer: Bob lashes out at Alice, but misjudges his attack and injures himself.



`)
}

func main() {
	flag.Usage = Usage
	var (
		a = flag.String("a", "m", "gender of attacker M, F, N or I")
		d = flag.String("d", "i", "gender of defender M, F, N or I")
		f = flag.String("f", "", "messge file to read, one per line")
		R = flag.Bool("R", false, "ommit raw message (bulk only)")
	)
	flag.Parse()

	switch *a {
	case "m", "M":
		*a = "MALE"
	case "f", "F":
		*a = "FEMALE"
	case "n", "N":
		*a = "NEUTRAL"
	case "i", "I":
		*a = "IT"
	default:
		fmt.Println("Attacker's gender (-a) must be one of M(ale), F(emale), N(eutral) or I(t).")
		os.Exit(-1)
	}

	switch *d {
	case "m", "M":
		*d = "MALE"
	case "f", "F":
		*d = "FEMALE"
	case "n", "N":
		*d = "NEUTRAL"
	case "i", "I":
		*d = "IT"
	default:
		fmt.Println("Defenser's gender (-d) must be one of M(ale), F(emale), N(eutral) or I(t).")
		os.Exit(-1)
	}

	e := entities()
	var msgs []string

	if len(flag.Args()) != 0 {
		msgs = []string{strings.Join(flag.Args(), " ")}
	}

	if len(msgs) == 0 {
		var (
			d   []byte
			err error
		)
		if *f != "" {
			d, err = os.ReadFile(*f)
			if err != nil {
				fmt.Println("Error reading stdin: %s", err)
				os.Exit(-1)
			}
		} else {
			d, err = io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Println("Error reading stdin: %s", err)
				os.Exit(-1)
			}
		}
		msgs = strings.Split(strings.TrimSpace(string(d)), "\n")
	}

	for x, msg := range msgs {
		am, dm, om := core.Message(e[*a], e[*d], msg)
		if !*R && len(msgs) > 1 {
			fmt.Printf("\n%17s: %s", fmt.Sprintf("raw message %d", x+1), msg)
		}
		if len(msgs) > 1 {
			fmt.Println()
		}
		fmt.Printf("%7s, attacker: %s\n", e[*a].As[core.UName], am)
		fmt.Printf("%7s, defender: %s\n", e[*d].As[core.UName], dm)
		fmt.Printf("%7s  observer: %s\n", "", om)
	}
}

func entities() map[string]*core.Thing {
	return map[string]*core.Thing{
		"FEMALE":  entity("Alice", "FEMALE"),
		"MALE":    entity("Bob", "MALE"),
		"NEUTRAL": entity("Charlie", "NEUTRAL"),
		"IT":      entity("an imp", "IT"),
	}
}

func entity(name, gender string) *core.Thing {
	r := recordjar.Record{
		"NAME":   []byte(name),
		"GENDER": []byte(gender),
	}
	t := core.NewThing()
	t.Unmarshal(r)
	return t
}
