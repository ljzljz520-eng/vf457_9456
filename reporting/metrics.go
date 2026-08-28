package reporting

import "streetlight/domain"

type CrewLoad struct {
	CrewID         string
	ActiveOrders   int
	AcceptedOrders int
}

func CrewLoads(crews []domain.Crew, orders []domain.WorkOrder) []CrewLoad {
	result := make([]CrewLoad, 0, len(crews))
	for _, crew := range crews {
		load := CrewLoad{CrewID: crew.ID}
		for _, order := range orders {
			if order.AssignedCrew != crew.ID {
				continue
			}
			if order.Status == domain.StatusAccepted {
				load.AcceptedOrders++
			} else {
				load.ActiveOrders++
			}
		}
		result = append(result, load)
	}
	return result
}
