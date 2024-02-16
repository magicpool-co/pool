package merkle

import (
	"github.com/magicpool-co/pool/pkg/crypto"
)

type Node struct {
	Left  *Node
	Right *Node
	Data  []byte
}

func newNode(left, right *Node, data []byte) *Node {
	node := Node{}
	if left == nil && right == nil {
		node.Data = crypto.ReverseBytes(data)
	} else {
		node.Data = crypto.Sha256d(append(left.Data, right.Data...))
	}

	node.Left = left
	node.Right = right

	return &node
}

func CalculateRoot(items [][]byte) []byte {
	if len(items) == 1 {
		// if there is one item, return that
		return items[0]
	} else if len(items)%2 != 0 {
		// if there are an odd number of items, duplicate the final item
		items = append(items, items[len(items)-1])
	}

	var nodes []*Node
	for _, item := range items {
		node := newNode(nil, nil, item)
		nodes = append(nodes, node)
	}

	for i := 0; i < len(items)/2; i++ {
		var level []*Node
		for j := 0; j < len(nodes); j += 2 {
			if len(nodes) < j+2 {
				nodes = append(nodes, nodes[j])
			}

			node := newNode(nodes[j], nodes[j+1], nil)
			level = append(level, node)
		}

		nodes = level
		if len(nodes) == 1 {
			break
		}
	}

	return crypto.ReverseBytes(nodes[0].Data)
}
