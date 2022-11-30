package kas

import (
	"sort"
)

type blockCache struct {
	minScore uint64
	maxScore uint64
	idx      map[string]*Block
	queue    map[string]bool
}

func newBlockCache(minScore, maxScore uint64, blocks ...*Block) *blockCache {
	cache := &blockCache{
		minScore: minScore,
		maxScore: maxScore,
		idx:      make(map[string]*Block),
		queue:    make(map[string]bool),
	}

	for _, block := range blocks {
		cache.add(block)
	}

	return cache
}

func (cache *blockCache) get(hash string) *Block {
	return cache.idx[hash]
}

func (cache *blockCache) add(block *Block) {
	if _, ok := cache.idx[block.Hash]; ok {
		return
	}
	cache.idx[block.Hash] = block

	for _, hash := range block.ChildrenHashes {
		if _, ok := cache.idx[hash]; !ok {
			cache.queue[hash] = true
		}
	}

	for _, hash := range block.MergeSetBluesHashes {
		if _, ok := cache.idx[hash]; !ok {
			cache.queue[hash] = true
		}
	}
}

func (cache *blockCache) size() int {
	return len(cache.queue)
}

func (cache *blockCache) list() []string {
	var count int
	list := make([]string, len(cache.queue))
	for hash := range cache.queue {
		list[count] = hash
		count++
	}

	sort.Strings(list)

	cache.queue = make(map[string]bool)

	return list
}
