package editbuf

import (
	"fmt"
	"strings"
)

const (
	BLOCKLEN = 512
)

var blockLen = BLOCKLEN // Exposed so tests can change this.

// left, buf, right
type node struct { // Should I store q0 at each span? lets me hold a cache of current page.
	left, right *node
	leftlength int
	buf []rune // If I have children I have no content.
	style Style
}

type Style interface {
}

type Editbuf struct {
	root *node
	len int
	blocklen int
}

func New() *Editbuf {
	return &Editbuf{
			root: newNode(nil,nil),
			len: 0,
			blocklen: blockLen,
		}
}

func (eb *Editbuf)Insert(q0 int, s []rune) {
	eb.root.insert(q0, s)
}

func (eb *Editbuf)String() string {
	treelen := eb.root.len()
	acc := make([]rune, 0, treelen)
	eb.root.string(0, treelen, &acc)
	return string(acc)
}


func (n *node)String() string {
	if n.buf != nil {
		return fmt.Sprintf("('%v'[%d])", string(n.buf), cap(n.buf))
	}
	b := strings.Builder{}
	b.WriteString("(")
	if n.left != nil {
		fmt.Fprintf(&b, "%v", n.left)
	}
	if n.right != nil {
		fmt.Fprintf(&b, "%v", n.right)
	}
	b.WriteString(")")
	return b.String()
}

func (n *node)len() (l int) {
	if n.buf != nil {
		return len(n.buf)
	}
	if n.right != nil {
		l = n.right.len()
	}
	return n.leftlength + l
}

// Insert allocates at most one buffer and two nodes to do the insertion.
// If the node is empty, it gets a copy of s.
// if the node would overflow inserting s, both sides of q0 are pushed
// into left and right and the node gets a copy of s.
// Returns the new root (in case of rebalancing)

func (n *node)insert(q0 int, s []rune) *node {
	// Do I belong left or right?
	if n.left != nil && q0 <= n.leftlength {
		n.leftlength += len(s)
		n.left = n.left.insert(q0, s)
		return n
	}
	if n.right != nil && q0 > n.leftlength {
		n.right = n.right.insert(q0 - n.leftlength, s)
		return n
	}
	// Do I fit in this node?
	if len(s) < cap(n.buf) - len(n.buf) {
		olen := len(n.buf)
		n.buf = n.buf[0:olen+len(s)]
		copy(n.buf[q0 + len(s):], n.buf[q0:olen]) // shift elements up
		copy(n.buf[q0:], s)	// Copy in s
		return n
	} 
	// Didn't fit.  Split and recurse.
	n.left = &node{nil, nil, 0, n.buf, nil}
	n.right = newNode(nil, nil) // &node{nil, nil, 0, nil, nil}
	if len(n.buf) > 0 {
		n.left.buf = n.buf[0:q0] // Re-use the buffer
		n.leftlength = q0
		n.right = n.right.insert(0, n.buf[q0:])
	}
	n.buf = nil
	if q0 < cap(n.left.buf)/2 { // fill the emptier side
		ilen := cap(n.left.buf) - len(n.left.buf)
		if ilen > len(s) {
			ilen = len(s)
		}
		n.left.buf = append(n.left.buf, s[0:ilen]...)
		n.left = n.left.insert(q0+ilen, s[ilen:]) // Or should I restart from the root to balance?
		n.leftlength += ilen
		return n
	} else {
		n.right = n.right.insert(0, s) 
		return n
	}
}

func newNode(left, right *node) *node{
	n := &node{
		left: left,
		right: right,
		buf: make([]rune, 0, blockLen),
		}
	if left != nil {
		n.leftlength = left.len()
	}
	return n
}

func (n *node)find(offset int) (*node, int) {
	if offset < n.leftlength {
		c, localoffset := n.left.find(offset)
		return c, localoffset
	} else if offset - n.leftlength < len(n.buf) {
		return n, offset - n.leftlength
	} else {
		c, localoffset := n.right.find(offset - n.leftlength - len(n.buf))
		return c, localoffset
	}
}

// To avoid allocations pass acc at least q1-q0 capacity
func (n *node)string(q0, q1 int, acc *[]rune) (used int) {
	if n.buf != nil {
		if q1 > len(n.buf) {
			q1 = len(n.buf)
		}
		*acc = append(*acc, n.buf[q0:q1]...)
		return q1-q0
	}
	if q0 < n.leftlength {
		used += n.left.string(q0, q1, acc)
		if used == q1 - q0 {
			return used
		}
		q0 = 0 // Make q0 & q1 relative to n.buf
		q1 -= used
	}
	// And down the right 
	used += n.right.string(0, q1 - q0, acc)
	return used
}
