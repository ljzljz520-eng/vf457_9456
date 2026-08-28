package dispatch

import (
	"sort"
	"streetlight/domain"
)

type Queue struct{ items []PlanItem }

func NewQueue() *Queue { return &Queue{items: make([]PlanItem, 0)} }

func (q *Queue) Add(item PlanItem) {
	q.items = append(q.items, item)
	sort.SliceStable(q.items, func(i, j int) bool {
		if q.items[i].Score == q.items[j].Score {
			return q.items[i].OrderID < q.items[j].OrderID
		}
		return q.items[i].Score > q.items[j].Score
	})
}

func (q *Queue) Pop() (PlanItem, bool) {
	if len(q.items) == 0 {
		return PlanItem{}, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

func (q *Queue) Len() int { return len(q.items) }

func BuildFilter(status domain.Status, circuit string, minimum int) domain.OrderFilter {
	return domain.OrderFilter{Status: status, CircuitID: circuit, MinimumPriority: minimum}
}
