package operations

import (
	"fmt"
	"sort"
	"streetlight/domain"
)

type AssetHealth struct {
	AssetID    string
	AssetType  string
	Condition  domain.AssetCondition
	OpenFaults int
}

type CircuitStatus struct {
	Circuit domain.CircuitLine
	Poles   []domain.StreetlightPole
	Boxes   []domain.ControlBox
	Health  domain.AssetCondition
}

func (s *Service) SetCircuitFault(circuitID, note string) (domain.CircuitLine, error) {
	circuit, err := s.DB.GetCircuitLine(circuitID)
	if err != nil {
		return domain.CircuitLine{}, err
	}
	if note == "" {
		return domain.CircuitLine{}, fmt.Errorf("fault note is required")
	}
	circuit.Faulted = true
	circuit.FaultNote = domain.NormalizeDescription(note)
	if circuit.FaultNote == "" {
		return domain.CircuitLine{}, fmt.Errorf("fault note is empty")
	}
	if err = s.DB.SaveCircuitLine(circuit); err != nil {
		return domain.CircuitLine{}, err
	}
	return circuit, nil
}

func (s *Service) ClearCircuitFault(circuitID string) (domain.CircuitLine, error) {
	circuit, err := s.DB.GetCircuitLine(circuitID)
	if err != nil {
		return domain.CircuitLine{}, err
	}
	circuit.Faulted = false
	circuit.FaultNote = ""
	if err = s.DB.SaveCircuitLine(circuit); err != nil {
		return domain.CircuitLine{}, err
	}
	return circuit, nil
}

func (s *Service) SetControlBoxOperational(boxID string, operational bool) (domain.ControlBox, error) {
	box, err := s.DB.GetControlBox(boxID)
	if err != nil {
		return domain.ControlBox{}, err
	}
	box.Operational = operational
	if err = s.DB.SaveControlBox(box); err != nil {
		return domain.ControlBox{}, err
	}
	return box, nil
}

func (s *Service) SetPoleActive(poleID string, active bool) (domain.StreetlightPole, error) {
	pole, err := s.DB.GetStreetlightPole(poleID)
	if err != nil {
		return domain.StreetlightPole{}, err
	}
	pole.Active = active
	if err = s.DB.SaveStreetlightPole(pole); err != nil {
		return domain.StreetlightPole{}, err
	}
	return pole, nil
}

func (s *Service) FindPoleByCode(code string) (domain.StreetlightPole, error) {
	if code == "" {
		return domain.StreetlightPole{}, fmt.Errorf("pole code is required")
	}
	poles, err := s.DB.ListStreetlightPoles()
	if err != nil {
		return domain.StreetlightPole{}, err
	}
	for _, pole := range poles {
		if pole.Code == code {
			return pole, nil
		}
	}
	return domain.StreetlightPole{}, fmt.Errorf("pole code %s not found", code)
}

func (s *Service) CircuitStatus(circuitID string) (CircuitStatus, error) {
	circuit, err := s.DB.GetCircuitLine(circuitID)
	if err != nil {
		return CircuitStatus{}, err
	}
	poles, err := s.DB.ListStreetlightPoles()
	if err != nil {
		return CircuitStatus{}, err
	}
	boxes, err := s.DB.ListControlBoxes()
	if err != nil {
		return CircuitStatus{}, err
	}
	poleIDs := make(map[string]bool)
	for _, id := range circuit.PoleIDs {
		poleIDs[id] = true
	}
	selectedPoles := make([]domain.StreetlightPole, 0)
	for _, pole := range poles {
		if poleIDs[pole.ID] {
			selectedPoles = append(selectedPoles, pole)
		}
	}
	selectedBoxes := make([]domain.ControlBox, 0)
	for _, box := range boxes {
		if box.CircuitID == circuitID {
			selectedBoxes = append(selectedBoxes, box)
		}
	}
	return CircuitStatus{Circuit: circuit, Poles: selectedPoles, Boxes: selectedBoxes, Health: domain.AssetConditionForCircuit(circuit)}, nil
}

func (s *Service) AssetHealthReport() ([]AssetHealth, error) {
	inv, err := s.DB.Inventory()
	if err != nil {
		return nil, err
	}
	result := make([]AssetHealth, 0, len(inv.Poles)+len(inv.Circuits)+len(inv.Orders))
	for _, pole := range inv.Poles {
		openFaults := 0
		for _, order := range inv.Orders {
			if order.PoleID == pole.ID && order.IsOpen() {
				openFaults++
			}
		}
		result = append(result, AssetHealth{AssetID: pole.ID, AssetType: "pole", Condition: domain.AssetConditionForPole(pole, inv.ControlBoxes), OpenFaults: openFaults})
	}
	for _, circuit := range inv.Circuits {
		openFaults := 0
		for _, order := range inv.Orders {
			if order.CircuitID == circuit.ID && order.IsOpen() {
				openFaults++
			}
		}
		result = append(result, AssetHealth{AssetID: circuit.ID, AssetType: "circuit", Condition: domain.AssetConditionForCircuit(circuit), OpenFaults: openFaults})
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].OpenFaults == result[j].OpenFaults {
			return result[i].AssetID < result[j].AssetID
		}
		return result[i].OpenFaults > result[j].OpenFaults
	})
	return result, nil
}

func (s *Service) FaultedPoleIDs(circuitID string) ([]string, error) {
	status, err := s.CircuitStatus(circuitID)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0)
	for _, pole := range status.Poles {
		if domain.AssetConditionForPole(pole, status.Boxes) != domain.ConditionOperational {
			result = append(result, pole.ID)
		}
	}
	sort.Strings(result)
	return result, nil
}

func (s *Service) ValidateAssetNetwork() error {
	if err := s.DB.ValidateReferences(); err != nil {
		return err
	}
	inv, err := s.DB.Inventory()
	if err != nil {
		return err
	}
	poles := make(map[string]bool)
	for _, pole := range inv.Poles {
		poles[pole.ID] = true
	}
	for _, circuit := range inv.Circuits {
		for _, poleID := range circuit.PoleIDs {
			if !poles[poleID] {
				return fmt.Errorf("circuit %s references unknown pole %s", circuit.ID, poleID)
			}
		}
	}
	return nil
}
