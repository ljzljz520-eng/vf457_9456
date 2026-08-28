package dispatch

import "streetlight/domain"

type Availability struct {
	CrewID     string
	OpenOrders int
	SkillCount int
}

func RankAvailability(crews []domain.Crew, orders []domain.WorkOrder) []Availability {
	result := make([]Availability, 0, len(crews))
	for _, crew := range crews {
		if !crew.Active {
			continue
		}
		open := 0
		for _, order := range orders {
			if order.AssignedCrew == crew.ID && order.Status != domain.StatusAccepted && order.Status != domain.StatusRejected {
				open++
			}
		}
		result = append(result, Availability{CrewID: crew.ID, OpenOrders: open, SkillCount: len(crew.Skills)})
	}
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].OpenOrders < result[i].OpenOrders {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}
