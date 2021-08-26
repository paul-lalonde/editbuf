package editbuf

import (
	"fmt"
	"runtime/debug"
	"testing"
)
/*
func TestNewBuf(t *testing.T) {
	buf := New()

	teststring := "This is a test"

	buf.Insert(0, []rune(teststring))

	s := buf.String()

	if s != teststring {
		t.Error("did not extract string")
	}
}
*/

func doubleInsert(n *node, or []rune, q0 int, r []rune) (nn *node, rs []rune) {
	rs = append(or[0:q0], append(r, or[q0:]...)...)
	nn = n.insert(q0, r)
	return nn, rs
}

func TestNode(t *testing.T) {

	type Test struct {
		setup func() (n *node, compare []rune)
		begin, end int
		expect []rune
		name string
	}

	testfunc := func(tests []Test) {
		acc := make([]rune, 0)
		var n *node
		for i, test := range tests {
			func() {
				defer func() {
					if err := recover(); err != nil {
						t.Errorf("Failed test %s: Panic: %s", test.name, err)
						debug.PrintStack()
						fmt.Print(n)
					}
				}()
				acc = acc[0:0]
fmt.Printf("%d: ", i)
				n, comp := test.setup()
fmt.Println(n, string(comp))
				used := n.string(test.begin, test.end, &acc)
				comp = comp[test.begin:test.end]
fmt.Println(n)
				if used != len(test.expect) || string(comp) != string(acc) {
					t.Errorf("Failed test %s: used = %d, expected %d; acc = %s, expected %s\n", test.name, used, test.end - test.begin, string(acc), string(comp))
				}
			}()
		}
	}

	text := []rune("Hello, 世界")
	lefttext := []rune("abcdef")
	righttext := []rune("uvwxyz")
	inserttext := []rune("Here I am")
	anothertext := []rune("01234")
	simplesetup := func() (*node, []rune){
		oblen := blockLen
		blockLen = 8
		n := newNode(nil, nil)
		n = n.insert(0, text)
		blockLen = oblen
		return n, text
	}
	fancysetup := func() (*node, []rune) {
		// Add left and right children
		left := newNode(nil, nil)
		left = left.insert(0, lefttext)
fmt.Println(left)
		right := newNode(nil, nil)
		right = right.insert(0, righttext)
fmt.Println(right)
		n := &node{left, right, left.len(), nil, nil}
fmt.Println(n)
		n = n.insert(len(lefttext), text)
fmt.Println(n)
		return n, append(lefttext, append(text, righttext...)...)
	}
	tests := []Test{
		{ simplesetup, 0, len(text), text, "trivial" },
		{ simplesetup, 1, len(text)-1, text[1:len(text)-1], "extract center from trivial" },
		//{ simplesetup, 0, len(text)+1, text, "Out of bounds"}, // Todo: this panics. We need to clean up.
		{ fancysetup, 0, 6, lefttext, "left child text" },
		{ fancysetup, 0, 5, lefttext[0:5], "left child left text" },
		{ fancysetup, 1, 5, lefttext[1:5], "left child center text" },
		{ func() (*node, []rune) {
				n, comp := simplesetup()
				n, comp = doubleInsert(n, comp, 0, inserttext)
				return n, comp
			},  
			0, len(inserttext)+len(text), append(inserttext, text...), "insert at zero" },
		{ func() (*node, []rune) {
				n, comp := simplesetup()
				n, comp = doubleInsert(n, comp, 1, inserttext)
fmt.Println("A:", n, string(comp))
				return n, comp
			}, 0, len(inserttext)+len(text), append(text[0:1], append(inserttext, text[1:]...)...), "insert at 1"},
		{ func() (*node, []rune) {
				n, comp := simplesetup()
				n, comp = doubleInsert(n, comp, 1, inserttext)
				n, comp = doubleInsert(n, comp, 0, anothertext)
				n, comp = doubleInsert(n, comp, 2, []rune("HHHH"))
				return n, comp
			}, 0, len(inserttext)+len(text), append(text[0:1], append(inserttext, text[1:]...)...), "insert at 1"},

		// Test that inserts to the left fill the root if they fit.
		
	}
	testfunc(tests)
}