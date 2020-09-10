// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package tree_test

import (
	"fmt"

	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// An example of using the tree package to produce the graphs in the package
// description.
func Example() {
	graph := tree.Tree{}
	graph.Append("Start").Append("Middle").Branch().Append("Second Branch").Append("With Text")
	graph.Append("End")

	fmt.Println(graph.Render())
	graph.Style = tree.StyleUnicode
	fmt.Println(graph.Render())
	// Output:
	//
	// |- Start
	// |- Middle
	// |  |- Second Branch
	// |  `- With Text
	// `- End
	//
	// ├─ Start
	// ├─ Middle
	// │  ├─ Second Branch
	// │  └─ With Text
	// └─ End
}

// A simple example of appending text nodes to a graph.
func ExampleNode_Append() {
	graph := tree.Tree{}
	graph.Style = tree.StyleUnicode
	graph.Append("First").Append("Second")
	fmt.Println(graph.Render())

	// Output:
	//
	// ├─ First
	// └─ Second
}

// A simple example of appending text nodes and a branch to a graph.
func ExampleNode_Branch() {
	graph := tree.Tree{}
	graph.Style = tree.StyleUnicode
	graph.Append("First").Branch().Append("Second")
	fmt.Println(graph.Render())

	// Output:
	//
	// └─ First
	//    └─ Second
}

// An example of using Tree.Indent when rendering. A custom style based on
// StyleUnicode has been used to make the whitespace visible as "·". The Indent
// controls the indentation of the graph from the left hand side. The Indent is
// independent of the tree.Width of the graph. The total width of the graph
// would therefore be Tree.Indent + Tree.Width when rendered.
func ExampleTree_indent() {
	graph := tree.Tree{}
	graph.Width = 30
	graph.Style = tree.StyleUnicode + "·"

	node := graph.Append("This is some short example text.")
	node = node.Branch()
	node = node.Append("This is some short example text.")

	graph.Indent = 0
	fmt.Println(graph.Render())
	graph.Indent = 2
	fmt.Println(graph.Render())
	graph.Indent = 4
	fmt.Println(graph.Render())

	// Output:
	// └─·This is some short example
	// ···text.
	// ···└─·This is some short
	// ······example text.
	//
	// ··└─·This is some short example
	// ·····text.
	// ·····└─·This is some short
	// ········example text.
	//
	// ····└─·This is some short example
	// ·······text.
	// ·······└─·This is some short
	// ··········example text.
}

// An example of using Tree.Offset when rendering. A custom style based on
// StyleUnicode has been used to make the whitespace visible as "·". The offset
// controls the amount of padding applied to continuation lines with respect to
// the initial line of a text node. Note that if the offset is zero and there
// is a branch node the branch will only be drawn after the continuation lines
// and not next to them.
func ExampleTree_offset() {
	graph := tree.Tree{}
	graph.Width = 30
	graph.Style = tree.StyleUnicode + "·"

	node := graph.Append("This is some short example text.")
	node = node.Branch()
	node = node.Append("This is some short example text.")

	graph.Offset = 0
	fmt.Println(graph.Render())
	graph.Offset = 2
	fmt.Println(graph.Render())
	graph.Offset = 8
	fmt.Println(graph.Render())

	// Output:
	//
	// └─·This is some short example
	// ···text.
	// ···└─·This is some short
	// ······example text.
	//
	// └─·This is some short example
	// ···│·text.
	// ···└─·This is some short
	// ········example text.
	//
	// └─·This is some short example
	// ···│·······text.
	// ···└─·This is some short
	// ··············example text.
}

// An example of using Tree.Width when rendering.
func ExampleTree_width() {
	graph := tree.Tree{}
	graph.Style = tree.StyleUnicode

	node := graph.Append("This is some short example text.")
	node = node.Branch()
	node = node.Append("This is some short example text.")

	graph.Width = 40
	fmt.Println(graph.Render())
	graph.Width = 30
	fmt.Println(graph.Render())
	graph.Width = 20
	fmt.Println(graph.Render())

	// Output:
	//
	// └─ This is some short example text.
	//    └─ This is some short example text.
	//
	// └─ This is some short example
	//    text.
	//    └─ This is some short
	//       example text.
	//
	// └─ This is some
	//    short example
	//    text.
	//    └─ This is some
	//       short example
	//       text.
}

// An example of using Tree.Offset when rendering with a fixed length prefix.
// In this case the prefix is a memory address. The tree.Offset value of 15 is
// derived from 12 positions for the address plus three positions for the
// space, hyphen and space that follow it.
func ExampleTree_Render_prefix() {
	graph := tree.Tree{}
	graph.Width = 40
	graph.Style = tree.StyleUnicode

	node := graph.Append("0xc000108000 - This is some short example text.")
	node = node.Append("0xc000108010 - This is some short example text.")
	node = node.Branch()
	node = node.Append("0xc000108020 - This is some short example text.")
	node = node.Append("0xc000108020 - This is some short example text.")
	node = node.Branch()
	node = node.Append("0xc000108030 - This is some short example text.")

	graph.Offset = 15
	fmt.Println(graph.Render())

	// Output:
	//
	// ├─ 0xc000108000 - This is some short
	// │                 example text.
	// └─ 0xc000108010 - This is some short
	//    │              example text.
	//    ├─ 0xc000108020 - This is some short
	//    │                 example text.
	//    └─ 0xc000108020 - This is some short
	//       │              example text.
	//       └─ 0xc000108030 - This is some
	//                         short example
	//                         text.
}

// A longer, contrived example of programmatically building a tree graph.
func Example_building() {

	bibliography := []struct {
		title     string
		authors   []string
		publisher string
		published string
		pages     int
		ISBN10    string
	}{
		{
			"The Go Programming Language",
			[]string{"Alan A. A. Donovan", "Brian W. Kernighan"},
			"Addison-Wesley Professional", "16 November 2015", 398, "9780134190440",
		},
		{
			"Programming in Go",
			[]string{"Mark Summerfield"},
			"Addison-Wesley", "30 April 2012", 496, "9780132764094",
		},
		{
			"Unix Text Processing",
			[]string{"Dale Dougherty", "Tim O'Reilly"},
			"Hayden Books", "1987", 665, "0672462915",
		},
	}

	graph := tree.Tree{}
	books := graph.Append("Book Catalogue").Branch()

	for _, info := range bibliography {
		books = books.Append(info.title)
		book := books.Branch()
		auth := book.Append("Author(s):").Branch()
		book.Append("Publisher: %s", info.publisher)
		book.Append("Published: %s", info.published)
		book.Append("Pages    : %d", info.pages)
		book.Append("ISBN-10  : %s", info.ISBN10)

		for _, author := range info.authors {
			auth.Append(author)
		}
	}

	graph.Style = tree.StyleUnicode
	fmt.Println(graph.Render())

	// Output:
	//
	// └─ Book Catalogue
	//    ├─ The Go Programming Language
	//    │  ├─ Author(s):
	//    │  │  ├─ Alan A. A. Donovan
	//    │  │  └─ Brian W. Kernighan
	//    │  ├─ Publisher: Addison-Wesley Professional
	//    │  ├─ Published: 16 November 2015
	//    │  ├─ Pages    : 398
	//    │  └─ ISBN-10  : 9780134190440
	//    ├─ Programming in Go
	//    │  ├─ Author(s):
	//    │  │  └─ Mark Summerfield
	//    │  ├─ Publisher: Addison-Wesley
	//    │  ├─ Published: 30 April 2012
	//    │  ├─ Pages    : 496
	//    │  └─ ISBN-10  : 9780132764094
	//    └─ Unix Text Processing
	//       ├─ Author(s):
	//       │  ├─ Dale Dougherty
	//       │  └─ Tim O'Reilly
	//       ├─ Publisher: Hayden Books
	//       ├─ Published: 1987
	//       ├─ Pages    : 665
	//       └─ ISBN-10  : 0672462915
}
