package storage

import (
	"sort"
	"streetlight/domain"
)

func SortWorkOrdersByPriority(items []domain.WorkOrder) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Priority == items[j].Priority {
			return items[i].ID < items[j].ID
		}
		return items[i].Priority > items[j].Priority
	})
}

func FindOrdersByCircuit(items []domain.WorkOrder, circuitID string) []domain.WorkOrder {
	out := make([]domain.WorkOrder, 0)
	for _, item := range items {
		if item.CircuitID == circuitID {
			out = append(out, item)
		}
	}
	SortWorkOrdersByPriority(out)
	return out
}

func FindOrdersByStatus(items []domain.WorkOrder, status domain.Status) []domain.WorkOrder {
	out := make([]domain.WorkOrder, 0)
	for _, item := range items {
		if item.Status == status {
			out = append(out, item)
		}
	}
	SortWorkOrdersByPriority(out)
	return out
}

func FindInspectionForOrder(items []domain.Inspection, orderID string) (domain.Inspection, bool) {
	var selected domain.Inspection
	found := false
	for _, item := range items {
		if item.WorkOrderID == orderID && (!found || item.Sequence > selected.Sequence) {
			selected, found = item, true
		}
	}
	return selected, found
}
