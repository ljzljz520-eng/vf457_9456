package workflow

import (
	"streetlight/domain"
	"streetlight/operations"
)

func Seed(service *operations.Service) error {
	return service.ImportInitialData(operations.InitialData{
		Circuits: []domain.CircuitLine{{ID: "circuit-1", Name: "North feeder", Voltage: 220, PoleIDs: []string{"pole-1", "pole-2"}}},
		Poles:    []domain.StreetlightPole{{ID: "pole-1", Code: "N-001", Location: "North Gate", LampType: "LED", Active: true}, {ID: "pole-2", Code: "N-002", Location: "North Park", LampType: "LED", Active: true}},
		Boxes:    []domain.ControlBox{{ID: "box-1", PoleID: "pole-1", CircuitID: "circuit-1", Address: "North Gate cabinet", Operational: true}},
		Crews:    []domain.Crew{{ID: "crew-1", Name: "Night electrical", Members: []string{"Ava", "Bo"}, Skills: []string{"cabinet", "LED"}, Active: true}},
	})
}
