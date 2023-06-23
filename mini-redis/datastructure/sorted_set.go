package datastructure

import "sort"

type SortedSet struct {
	items []SortedSetItem
	index map[string]int
}

func MakeSortedSet() *SortedSet {
	return &SortedSet{
		items: make([]SortedSetItem, 0),
		index: make(map[string]int),
	}
}

type SortedSetItem struct {
	Score  float64
	Member string
}

func (set *SortedSet) Set(score float64, member string) bool {
	item := SortedSetItem{score, member}
	defer set.ensureOrder()

	if index, ok := set.index[member]; ok {
		set.items[index] = item
		return false
	}

	set.items = append(set.items, item)
	return true
}

func (set *SortedSet) ensureOrder() {
	sort.SliceStable(set.items, func(i, j int) bool {
		return set.items[i].Score < set.items[j].Score
	})

	for index, item := range set.items {
		set.index[item.Member] = index
	}
}

func (set *SortedSet) Len() int {
	return len(set.items)
}
