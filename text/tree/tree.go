// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package tree is used to produce textual tree graphs. The graph can be
// produced using ASCII, Unicode or custom characters:
//
//               |- Start                   ├─ Start
//               |- Middle                  ├─ Middle
//               |  |- Second Branch        │  ├─ Second Branch
//               |  `- With Text            │  └─ With Text
//               `- End                     └─ End
//
// The zero value of a Tree is ready to be populated by appending text and
// branch nodes using the Append and Branch methods. Once all nodes have been
// added the tree can be rendered by calling the Render method.
//
// The maximum depth and breadth of the tree is very large, usually only limited
// by the memory available.
package tree

import (
	"fmt"
	"strings"
)

// Predefined style constants for rendering the tree output. The Tree.Style
// field can be set to one of the predefined style strings or a custom string
// of the form "vhjes" where:
//
//  v - vertical branch
//  h - horizontal branch
//  j - join between a vertical and horizontal branch
//  e - the elbow at the end of a branch
//  s - character for indentation and gaps between branches
//
// For example, the following is rendered with the StyleUnicode "│─├└" string:
//
//  ├─   jhs
//  │    vss
//  └─   ehs
//
// If a custom style string is less than 5 characters long then spaces ' '
// (ASCII 0x20) will be used for the missing characters. If Tree.Style is the
// default empty string then StyleASCII is used.
const (
	StyleASCII       = "|-|`"
	StyleUnicode     = "│─├└"
	StyleUnicodeBold = "┃━┣┗"
)

// Tree is used to hold working data and configuration settings for a tree
// graph. The zero value of a Tree is ready to use. At least one text or branch
// node needs to be added to the Tree via Tree.Append or Tree.Branch initially
// in order to obtain a tree.Node instance. All other text and branch nodes can
// then be added using Node.Append or Node.Branch methods on the returned Node.
type Tree struct {

	// Width is the maximum width of rendered graph, excluding indentation. Text
	// nodes that exceed the width will be wrapped on to continuation lines. If
	// set to the default of 0 no wrapping occurs. Minimum width should be 3 *
	// max tree depth + a reasonable amount for text. If there is no way to
	// reasonably fit the text to the graph it will be displayed as is,
	// regardless of the Width setting. If negative will be set to 0 when the
	// graph is rendered.
	Width int

	// Indent is the amount to indent the rendered graph from the left hand side.
	// The Indent is independent of and not included in the Width. If negative
	// will be set to 0 when the graph is rendered.
	Indent int

	// Offset is the amount to offset continuation lines with respect to the
	// initial line. If set to the default of 0 then continuation lines will not
	// have branches drawn immediately next to them, but underneath. If negative
	// will be set to 0 when the graph is rendered.
	Offset int

	// Runes to render graph. See Style* constants for details and predefined
	// styles. If set to the default empty string it will be set to StyleASCII
	// when the graph is rendered.
	Style string

	root *Node // Root of linked *Node making up the current tree

	// Pre-calculated parts of the graph based on the current style
	item    []rune // "├─ "
	branch  []rune // "│  "
	end     []rune // "└─ "
	void    []rune // "   "
	padding string // padding for continuation line offsets

	branches []rune // Branches being rendered for current graph line
}

// Append adds a new text node to the end of the root branch and returns the
// new Node. The passed format and args parameters are processed by the
// fmt.Sprintf function.
//
// If the text passed to Append evaluates to the empty string no new text node
// will be created and the receiver node will be returned. However, if there
// are no other text nodes on the root branch this is the equivalent of calling
// the Tree.Branch method, which has no visible effect on the resulting graph
// but allows for a Node to be returned.
func (t *Tree) Append(format string, args ...interface{}) *Node {
	if t.root == nil {
		t.root = &Node{text: fmt.Sprintf(format, args...)}
		return t.root
	}
	return t.root.Append(format, args...)
}

// Branch adds a new branch node to the receiver node and returns the new node.
// If the receiver node already has a branch from it the branch is followed to
// its end, until a node without a branch is found, and the new branch is then
// made at the end.
func (t *Tree) Branch() *Node {
	if t.root == nil {
		t.root = &Node{text: ""}
		return t.root
	}
	return t.root.Branch()
}

// Node is used to represent the tree like structure of text and branch nodes.
// A Node may point to the next text node on the branch and/or to a different
// branch node.
//
// A new text node can be appended to a branch by calling the Append method. A
// new branch can be added using the Branch method.
type Node struct {
	text   string
	branch *Node
	next   *Node
}

// Append adds a new text node to the end of the current branch pointed to by
// the receiver and returns the new node. The passed format and args parameters
// are processed by the fmt.Sprintf function.
//
// If the text passed to Append evaluates to the empty string no new text node
// will be created and the receiver node will be returned.
//
// The returned node may be saved to bookmark a position in the tree. The saved
// node can then be used to add additional nodes and/or branches.
func (n *Node) Append(format string, args ...interface{}) *Node {
	text := fmt.Sprintf(format, args...)
	if text == "" {
		return n
	}
	for ; n.next != nil; n = n.next {
	}
	n.next = &Node{text, nil, nil}
	return n.next
}

// Branch adds a new branch node to the receiver node and returns the new node.
// If the receiver node already has a branch from it the branch is followed to
// its end, until a node without a branch is found, and the new branch is then
// made at the end.
//
// A new branch node will not be added to an empty branch - one that has not
// had any text nodes added to it. In this case Branch will return the empty
// branch node found.
//
// The returned Node may be saved to bookmark a position in the tree. The saved
// Node can then be used to add additional nodes and/or branches.
func (n *Node) Branch() *Node {
	for ; n.branch != nil; n = n.branch {
	}
	if n.text == "" && n.next == nil {
		return n
	}
	for ; n.next != nil; n = n.next {
	}

	// We need a node to link the branch to so we create a node with no text as
	// this cannot be created by a user. Nodes with empty text are not rendered.
	n.branch = &Node{"", nil, nil}

	return n.branch
}

// Render processes the current tree structure and returns the resulting graph
// as a string. The same Tree may be rendered multiple times using different
// settings. See the Tree type for configurable settings.
func (t *Tree) Render() (graph string) {

	if t.root == nil { // If tree has no nodes then nothing to do
		return ""
	}

	// Sanity checks
	if t.Width < 0 {
		t.Width = 0
	}
	if t.Indent < 0 {
		t.Indent = 0
	}
	if t.Offset < 0 {
		t.Offset = 0
	}
	if t.Style == "" {
		t.Style = StyleASCII
	}

	s := []rune(t.Style + "     ")
	vertical, horizontal, join, elbow, space := s[0], s[1], s[2], s[3], s[4]

	// Build tree drawing elements - saves on multiple appends in render
	t.item = []rune{join, horizontal, space}  // '├─ '
	t.branch = []rune{vertical, space, space} // '│  '
	t.end = []rune{elbow, horizontal, space}  // '└─ '
	t.void = []rune{space, space, space}      // '   '

	// Padding for continuation lines, minus 1 to fit a vertical or space rune
	if t.Offset > 0 {
		t.padding = strings.Repeat(string(space), t.Offset-1)
	}

	// Reset branches and setup initial branch. Setting up the initial branch
	// here which saves on corner-cases to deal with in render.
	t.branches = []rune(strings.Repeat(string(space), t.Indent))
	t.appendIf(t.root.next == nil, t.end, t.item)

	return string(t.render(t.root))
}

// render is the main, recursive method used to produce the graph for the
// current Tree.
func (t *Tree) render(n *Node) (graph []byte) {

	// Render current node if text node
	if n.text != "" {

		// Render initial line
		line, text := t.breakLine(n.text, t.Width+t.Indent-len(t.branches))
		graph = append(graph, string(t.branches)...)
		graph = append(graph, line...)
		graph = append(graph, '\n')

		// If more text remains then render continuation lines
		if text != "" {

			// Fixup previous branch
			t.replaceIf(n.next != nil, t.branch, t.void)

			// If we have space in the offset then render continuation of current branch
			if t.Offset > 0 {
				t.appendIf(
					n.branch != nil && n.branch.next != nil, t.branch[:1], t.void[:1],
				)
			}

			for text != "" {
				line, text = t.breakLine(text, t.Width-t.Offset+t.Indent-len(t.branches))
				graph = append(graph, string(t.branches)...)
				graph = append(graph, t.padding...)
				graph = append(graph, line...)
				graph = append(graph, '\n')
			}

			// If continuation of current branch rendered remove it again
			if t.Offset > 0 {
				t.branches = t.branches[:len(t.branches)-1]
			}
		}
	}

	// Render branch - fixup previous branch, add new branch and remove when done
	if n.branch != nil {
		t.replaceIf(n.next != nil, t.branch, t.void)
		t.appendIf(n.branch.next == nil, t.end, t.item)
		graph = append(graph, t.render(n.branch)...)
		t.branches = t.branches[:len(t.branches)-3]
	}

	// Render text node - either at the end of the branch or on it
	if n.next != nil {
		t.replaceIf(n.next.next == nil, t.end, t.item)
		graph = append(graph, t.render(n.next)...)
	}

	return
}

// appendIf is a helper method to add T to the end of branches if test is true
// else adds F. Eliminates many if..else statements and should get inlined.
func (t *Tree) appendIf(test bool, T, F []rune) {
	if test {
		t.branches = append(t.branches, T...)
		return
	}
	t.branches = append(t.branches, F...)
}

// replaceIf is a helper method to replace the last three branch runes with T
// if test is true else replaces with F. Eliminates many if..else statements
// and should get inlined.
func (t *Tree) replaceIf(test bool, T, F []rune) {
	if test {
		copy(t.branches[len(t.branches)-3:], T)
		return
	}
	copy(t.branches[len(t.branches)-3:], F)
}

// breakLine breaks the given text, on ASCII space if possible, to fit within
// the given width returning the line broken off and any remaining text. If
// width is zero or less the original text is returned. If there are no spaces
// to break on in the given text the text is broken after width and not on
// whitespace.
func (t *Tree) breakLine(text string, width int) (line, remain string) {
	if width <= 0 || len(text) <= width {
		return text, ""
	}
	if i := strings.LastIndex(text[:width], " "); i != -1 {
		return text[:i], text[i+1:]
	}
	return text[:width], text[width:]
}
