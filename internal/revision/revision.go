package revision

import (
	"cmp"
	"container/heap"
	"fmt"

	"github.com/tychonis/cyanotype/model"
)

// StableTopoRevisions returns a topological order where unrelated ready revisions
// are ordered by CreatedAt, then RevisionDigest.
func StableTopoRevisions(revisions []*model.Revision) ([]model.RevisionID, error) {
	byID := make(map[model.RevisionID]*model.Revision, len(revisions))
	children := make(map[model.RevisionID][]model.RevisionID, len(revisions))
	indegree := make(map[model.RevisionID]int, len(revisions))

	for _, r := range revisions {
		if r.Digest == "" {
			return nil, fmt.Errorf("empty revision ID")
		}
		if _, exists := byID[r.Digest]; exists {
			return nil, fmt.Errorf("duplicate revision ID %q", r.Digest)
		}
		byID[r.Digest] = r
		indegree[r.Digest] = 0
	}

	for _, r := range revisions {
		for _, p := range r.Parents {
			if _, exists := byID[p]; !exists {
				return nil, fmt.Errorf("revision %q has unknown parent %q", r.Digest, p)
			}
			children[p] = append(children[p], r.Digest)
			indegree[r.Digest]++
		}
	}

	pq := &revisionHeap{}
	heap.Init(pq)

	for _, r := range revisions {
		if indegree[r.Digest] == 0 {
			heap.Push(pq, r)
		}
	}

	order := make([]model.RevisionID, 0, len(revisions))

	for pq.Len() > 0 {
		r := heap.Pop(pq).(*model.Revision)
		order = append(order, r.Digest)

		for _, childID := range children[r.Digest] {
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
	return a.Digest < b.Digest
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
