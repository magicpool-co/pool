package util

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	if left == nil && right == nil {
		node.Data = ReverseBytes(data)
	} else {
		node.Data = Sha256d(append(left.Data, right.Data...))
	}

	node.Left = left
	node.Right = right

	return &node
}

func CalculateMerkleRoot(data [][]byte) []byte {
	if len(data) == 1 {
		return data[0]
	}

	var nodes []MerkleNode

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, dat := range data {
		node := NewMerkleNode(nil, nil, dat)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var level []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			if len(nodes) < j+2 {
				nodes = append(nodes, nodes[j])
			}

			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			level = append(level, *node)
		}

		nodes = level
		if len(nodes) == 1 {
			break
		}
	}

	return ReverseBytes(nodes[0].Data)
}
