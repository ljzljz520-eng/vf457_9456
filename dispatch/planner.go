package dispatch

import (
	"fmt"
	"streetlight/domain"
	"streetlight/operations"
)

type PlanItem struct {
	OrderID string
	CrewID  string
	Score   int
	Reason  string
}

type Planner struct{ Service *operations.Service }

func NewPlanner(service *operations.Service) *Planner { return &Planner{Service: service} }

func (p *Planner) Recommend(filter domain.OrderFilter) ([]PlanItem, error) {
	orders, err := p.Service.ListOrders(filter)
	if err != nil {
		return nil, err
	}
	crews, err := p.Service.DB.ListCrews()
	if err != nil {
		return nil, err
	}
	items := make([]PlanItem, 0)
	for _, order := range orders {
		crew, ok := chooseCrew(crews, order)
		if !ok {
			continue
		}
		items = append(items, PlanItem{OrderID: order.ID, CrewID: crew.ID, Score: order.Priority*10 + len(crew.Skills), Reason: "priority and skills matched"})
	}
	return items, nil
}

func chooseCrew(crews []domain.Crew, order domain.WorkOrder) (domain.Crew, bool) {
	var selected domain.Crew
	found := false
	for _, crew := range crews {
		if !crew.Active {
			continue
		}
		if order.AssignedCrew != "" && crew.ID == order.AssignedCrew {
			return crew, true
		}
		if !found || len(crew.Skills) > len(selected.Skills) || (len(crew.Skills) == len(selected.Skills) && crew.ID < selected.ID) {
			selected, found = crew, true
		}
	}
	return selected, found
}

func (p *Planner) Dispatch(item PlanItem, dispatcher string) error {
	if item.OrderID == "" || item.CrewID == "" {
		return fmt.Errorf("plan item is incomplete")
	}
	return p.Service.AssignCrew(domain.Assignment{WorkOrderID: item.OrderID, CrewID: item.CrewID, Dispatcher: dispatcher, Reason: item.Reason})
}

func (p *Planner) DispatchAll(items []PlanItem, dispatcher string) (int, error) {
	count := 0
	for _, item := range items {
		if err := p.Dispatch(item, dispatcher); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
