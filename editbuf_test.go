package editbuf

import (
	"runtime/debug"
	"testing"
)

func TestNewBuf(t *testing.T) {
	buf := New()

	teststring := "This is a test"

	buf.Insert(0, []rune(teststring))

	s := buf.String()

	if s != teststring {
		t.Error("did not extract string")
	}
}

func TestNode(t *testing.T) {

	type Test struct {
		setup func() *node
		begin, end int
		expect []rune
		name string
	}

	testfunc := func(tests []Test) {
		acc := make([]rune, 0)
		for _, test := range tests {
			func() {
				defer func() {
					if err := recover(); err != nil {
						t.Errorf("Failed test %s: Panic: %s", test.name, err)
						debug.PrintStack()
					}
				}()
				acc = acc[0:0]
				n := test.setup()
				used := n.string(test.begin, test.end, &acc)
				if used != len(test.expect) || string(test.expect) != string(acc) {
					t.Errorf("Failed test %s: used = %d, expected %d; acc = %s, expected %s\n", test.name, used, test.end - test.begin, string(acc), string(test.expect))
				}
			}()
		}
	}

	text := []rune("Hello, 世界")
	lefttext := []rune("abcdef")
	righttext := []rune("uvwxyz")
	inserttext := []rune("Here I am")
	simplesetup := func() *node {
		return newNode(nil, nil, text)
	}
	fancysetup := func() *node {
		// Add left and right children
		left := newNode(nil, nil, lefttext)
		right := newNode(nil, nil, righttext)
		return newNode(left, right, text)
		
	}
	tests := []Test{
		{ simplesetup, 0, len(text), text, "trivial" },
		{ simplesetup, 1, len(text)-1, text[1:len(text)-1], "extract center from trivial" },
		//{ simplesetup, 0, len(text)+1, text, "Out of bounds"}, // Todo: this panics. We need to clean up.
		{ fancysetup, 0, 6, lefttext, "left child text" },
		{ fancysetup, 0, 5, lefttext[0:5], "left child left text" },
		{ fancysetup, 1, 5, lefttext[1:5], "left child center text" },
		{ func() *node {
				n := simplesetup()
				return n.insert(0, inserttext)
			},  
			0, len(inserttext)+len(text), append(inserttext, text...), "insert at zero" },
		{ func() *node {
				n := simplesetup()
				return n.insert(1, inserttext)
			}, 0, len(inserttext)+len(text), append(text[0:1], append(inserttext, text[1:]...)...), "insert at 1"},
		
	}
	testfunc(tests)
}