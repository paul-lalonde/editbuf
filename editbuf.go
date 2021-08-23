package editbuf

// left, buf, right
type node struct {
	left, right *node
	leftlength, fulllength int
	buf []rune
	style Style
}

type Style interface {
}

type Editbuf struct {
	root *node
	len int
}

func New() *Editbuf {
	return &Editbuf{
		&node{nil, nil, 0, 0, make([]rune, 0, 128), nil}, 0,
	}
}

func (eb *Editbuf)Insert(q0 int, s []rune) {
	eb.root.insert(q0, s)
}

func (eb *Editbuf)String() string {
	acc := make([]rune, 0, eb.root.fulllength)
	eb.root.string(0, eb.root.fulllength, &acc)
	return string(acc)
}

// Insert allocates at most one buffer and two nodes to do the insertion.
// If the node is empty, it gets a copy of s.
// if the node would overflow inserting s, both sides of q0 are pushed
// into left and right and the node gets a copy of s.
// Returns the new root (in case of rebalancing)

func (n *node)insert(q0 int, s []rune) *node {
	// If I'm visiting this node, it and/or its children will get longer
	n.fulllength += len(s)
	if q0 < n.leftlength {
		n.left.insert(q0,s) // Always fully insert.
		return n
	} else if q0 < len(n.buf) {
		// insertion point is in my buffer
		n.buf = append(n.buf[0:q0 - n.leftlength], append(s, n.buf[q0:]...)...)
		// Missing logic - if I grow longer than my buffer I need to 
		// split at the insertion point, pushing part left, part right,
		// and leaving buf == s
		return n
	} else {
		// Insertion point is to my right
		if n.right == nil {
			n.right = newNode(nil, nil, s)
		} else {
			n.right.insert(q0 - n.leftlength - len(n.buf), s)
		}
		return n
	}
}

// Expensive, traverses the whole tree.
func (n *node) length() int {
	return n.fulllength
}

func newNode(left, right *node, s []rune) *node{
	n := &node{left, right, 0, len(s), s, nil}
	if left != nil {
		n.leftlength = left.length()
	}
	if right != nil {
		n.fulllength += right.fulllength
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
	checklen := q1 - q0
	if q0 < n.leftlength {
		used = n.left.string(q0, q1, acc)
		if used == q1 - q0 {
			return used
		}
		q0 = 0 // Make q0 & q1 relative to n.buf
		q1 -= used
	}
	if q1 <= len(n.buf) {
		*acc = append(*acc, n.buf[q0:q1]...)
		used += q1 - q0
		return used
	} 
	// And down the right 
	*acc = append(*acc, n.buf[q0:]...)
	used += len(n.buf)
	q1 -= len(n.buf)
	used += n.right.string(0, q1, acc)
	if used != checklen {
		panic("Failed to find all my segments")
	}
	return used
}
