package ranker

import (
	"cmp"
	"container/heap"
	"fmt"

	"github.com/tychonis/cyanotype/model"
)

// StableTopoRevisions returns a topological order where unrelated ready revisions
// are ordered by CreatedAt, then RevisionID.
func StableTopoRevisions(revisions []*model.Revision) ([]model.RevisionID, error) {
	byID := make(map[model.RevisionID]model.Revision, len(revisions))
	children := make(map[model.RevisionID][]model.RevisionID, len(revisions))
	indegree := make(map[model.RevisionID]int, len(revisions))

	for _, r := range revisions {
		if r.ID == "" {
			return nil, fmt.Errorf("empty revision ID")
		}
		if _, exists := byID[r.ID]; exists {
			return nil, fmt.Errorf("duplicate revision ID %q", r.ID)
		}
		byID[r.ID] = *r
		indegree[r.ID] = 0
	}

	for _, r := range revisions {
		for _, p := range r.Parents {
			if _, exists := byID[p]; !exists {
				return nil, fmt.Errorf("revision %q has unknown parent %q", r.ID, p)
			}
			children[p] = append(children[p], r.ID)
			indegree[r.ID]++
		}
	}

	pq := &revisionHeap{}
	heap.Init(pq)

	for _, r := range revisions {
		if indegree[r.ID] == 0 {
			heap.Push(pq, r)
		}
	}

	order := make([]model.RevisionID, 0, len(revisions))

	for pq.Len() > 0 {
		r := heap.Pop(pq).(*model.Revision)
		order = append(order, r.ID)

		for _, childID := range children[r.ID] {
			indegree[childID]--
			if indegree[childID] == 0 {
				heap.Push(pq, byID[childID])
			}
		}
	}

	if len(order) != len(revisions) {
		return nil, fmt.Errorf("cycle detected in revision graph")
	}

	return order, nil
}

type revisionHeap []*model.Revision

func (h revisionHeap) Len() int { return len(h) }

func (h revisionHeap) Less(i, j int) bool {
	a, b := h[i], h[j]

	if c := cmp.Compare(a.CreatedAt, b.CreatedAt); c != 0 {
		return c < 0
	}
	return a.ID < b.ID
}

func (h revisionHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *revisionHeap) Push(x any) {
	*h = append(*h, x.(*model.Revision))
}

func (h *revisionHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
