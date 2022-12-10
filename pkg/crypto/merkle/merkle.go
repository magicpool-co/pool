package merkle

import (
	"github.com/magicpool-co/pool/pkg/crypto"
)

type Node struct {
	Left  *Node
	Right *Node
	Data  []byte
}

func NewNode(left, right *Node, data []byte) *Node {
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

func CalculateRoot(data [][]byte) []byte {
	if len(data) == 1 {
		return data[0]
	}

	var nodes []Node

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, dat := range data {
		node := NewNode(nil, nil, dat)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var level []Node

		for j := 0; j < len(nodes); j += 2 {
			if len(nodes) < j+2 {
				nodes = append(nodes, nodes[j])
			}

			node := NewNode(&nodes[j], &nodes[j+1], nil)
			level = append(level, *node)
		}

		nodes = level
		if len(nodes) == 1 {
			break
		}
	}

	return crypto.ReverseBytes(nodes[0].Data)
}
