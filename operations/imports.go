package operations

import (
	"fmt"
	"streetlight/domain"
)

type InitialData struct {
	Poles    []domain.StreetlightPole
	Circuits []domain.CircuitLine
	Boxes    []domain.ControlBox
	Crews    []domain.Crew
}

func (s *Service) ImportInitialData(data InitialData) error {
	for _, circuit := range data.Circuits {
		if err := s.RegisterCircuit(circuit); err != nil {
			return fmt.Errorf("circuit %s: %w", circuit.ID, err)
		}
	}
	for _, pole := range data.Poles {
		if err := s.RegisterPole(pole); err != nil {
			return fmt.Errorf("pole %s: %w", pole.ID, err)
		}
	}
	for _, box := range data.Boxes {
		if err := s.RegisterControlBox(box); err != nil {
			return fmt.Errorf("box %s: %w", box.ID, err)
		}
	}
	for _, crew := range data.Crews {
		if err := s.RegisterCrew(crew); err != nil {
			return fmt.Errorf("crew %s: %w", crew.ID, err)
		}
	}
	return s.DB.ValidateReferences()
}
