package storage

import (
	"encoding/json"
	"fmt"
	"go.etcd.io/bbolt"
	"streetlight/domain"
)

type Snapshot struct {
	Poles        []domain.StreetlightPole `json:"poles"`
	ControlBoxes []domain.ControlBox      `json:"control_boxes"`
	Circuits     []domain.CircuitLine     `json:"circuits"`
	Crews        []domain.Crew            `json:"crews"`
	Orders       []domain.WorkOrder       `json:"orders"`
	Inspections  []domain.Inspection      `json:"inspections"`
	Repairs      []domain.RepairRecord    `json:"repairs"`
}

func (d *DB) CreateSnapshot() (Snapshot, error) {
	inv, err := d.Inventory()
	if err != nil {
		return Snapshot{}, err
	}
	inspections, err := d.ListInspections()
	if err != nil {
		return Snapshot{}, err
	}
	repairs, err := d.ListRepairRecords()
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Poles: inv.Poles, ControlBoxes: inv.ControlBoxes, Circuits: inv.Circuits, Crews: inv.Crews, Orders: inv.Orders, Inspections: inspections, Repairs: repairs}, nil
}

func (s Snapshot) Validate() error {
	if len(s.Poles) == 0 && len(s.Circuits) == 0 && len(s.Crews) == 0 && len(s.Orders) == 0 {
		return fmt.Errorf("snapshot has no core records")
	}
	ids := make(map[string]bool)
	for _, pole := range s.Poles {
		if pole.ID == "" || ids[pole.ID] {
			return fmt.Errorf("invalid or duplicate pole %s", pole.ID)
		}
		ids[pole.ID] = true
	}
	for _, circuit := range s.Circuits {
		if circuit.ID == "" {
			return fmt.Errorf("circuit id is required")
		}
	}
	for _, order := range s.Orders {
		if err := order.Validate(); err != nil {
			return err
		}
	}
	for _, inspection := range s.Inspections {
		if inspection.ID == "" || inspection.WorkOrderID == "" {
			return fmt.Errorf("inspection identity is incomplete")
		}
	}
	return nil
}

func (d *DB) SnapshotJSON() ([]byte, error) {
	snapshot, err := d.CreateSnapshot()
	if err != nil {
		return nil, err
	}
	if err = snapshot.Validate(); err != nil {
		return nil, err
	}
	return json.MarshalIndent(snapshot, "", "  ")
}

func DecodeSnapshot(data []byte) (Snapshot, error) {
	if len(data) == 0 {
		return Snapshot{}, fmt.Errorf("snapshot data is empty")
	}
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, err
	}
	if err := snapshot.Validate(); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func (d *DB) RestoreSnapshot(snapshot Snapshot) error {
	if err := snapshot.Validate(); err != nil {
		return err
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.db == nil {
		return fmt.Errorf("database is closed")
	}
	return d.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range bucketNames {
			if err := tx.Bucket(name).ForEach(func(key, _ []byte) error { return tx.Bucket(name).Delete(key) }); err != nil {
				return err
			}
		}
		write := func(bucket string, key string, value any) error {
			data, err := json.Marshal(value)
			if err != nil {
				return err
			}
			return tx.Bucket([]byte(bucket)).Put([]byte(key), data)
		}
		for _, pole := range snapshot.Poles {
			if err := write("streetlight_poles", pole.ID, pole); err != nil {
				return err
			}
		}
		for _, box := range snapshot.ControlBoxes {
			if err := write("control_boxes", box.ID, box); err != nil {
				return err
			}
		}
		for _, circuit := range snapshot.Circuits {
			if err := write("circuit_lines", circuit.ID, circuit); err != nil {
				return err
			}
		}
		for _, crew := range snapshot.Crews {
			if err := write("crews", crew.ID, crew); err != nil {
				return err
			}
		}
		for _, order := range snapshot.Orders {
			if err := write("work_orders", order.ID, order); err != nil {
				return err
			}
		}
		for _, inspection := range snapshot.Inspections {
			if err := write("inspections", inspection.ID, inspection); err != nil {
				return err
			}
		}
		for _, repair := range snapshot.Repairs {
			if err := write("repair_records", repair.ID, repair); err != nil {
				return err
			}
		}
		return nil
	})
}

func MergeSnapshots(base, incoming Snapshot) (Snapshot, error) {
	if err := base.Validate(); err != nil {
		return Snapshot{}, err
	}
	if err := incoming.Validate(); err != nil {
		return Snapshot{}, err
	}
	result := base
	result.Poles = mergePoles(base.Poles, incoming.Poles)
	result.ControlBoxes = mergeBoxes(base.ControlBoxes, incoming.ControlBoxes)
	result.Circuits = mergeCircuits(base.Circuits, incoming.Circuits)
	result.Crews = mergeCrews(base.Crews, incoming.Crews)
	result.Orders = mergeOrders(base.Orders, incoming.Orders)
	result.Inspections = mergeInspections(base.Inspections, incoming.Inspections)
	result.Repairs = mergeRepairs(base.Repairs, incoming.Repairs)
	return result, nil
}

func mergePoles(base, incoming []domain.StreetlightPole) []domain.StreetlightPole {
	result := append([]domain.StreetlightPole(nil), base...)
	index := make(map[string]int)
	for i, item := range result {
		index[item.ID] = i
	}
	for _, item := range incoming {
		if i, found := index[item.ID]; found {
			result[i] = item
		} else {
			index[item.ID] = len(result)
			result = append(result, item)
		}
	}
	return result
}

func mergeBoxes(base, incoming []domain.ControlBox) []domain.ControlBox {
	result := append([]domain.ControlBox(nil), base...)
	index := make(map[string]int)
	for i, item := range result {
		index[item.ID] = i
	}
	for _, item := range incoming {
		if i, found := index[item.ID]; found {
			result[i] = item
		} else {
			index[item.ID] = len(result)
			result = append(result, item)
		}
	}
	return result
}

func mergeCircuits(base, incoming []domain.CircuitLine) []domain.CircuitLine {
	result := append([]domain.CircuitLine(nil), base...)
	index := make(map[string]int)
	for i, item := range result {
		index[item.ID] = i
	}
	for _, item := range incoming {
		if i, found := index[item.ID]; found {
			result[i] = item
		} else {
			index[item.ID] = len(result)
			result = append(result, item)
		}
	}
	return result
}

func mergeCrews(base, incoming []domain.Crew) []domain.Crew {
	result := append([]domain.Crew(nil), base...)
	index := make(map[string]int)
	for i, item := range result {
		index[item.ID] = i
	}
	for _, item := range incoming {
		if i, found := index[item.ID]; found {
			result[i] = item
		} else {
			index[item.ID] = len(result)
			result = append(result, item)
		}
	}
	return result
}

func mergeOrders(base, incoming []domain.WorkOrder) []domain.WorkOrder {
	result := append([]domain.WorkOrder(nil), base...)
	index := make(map[string]int)
	for i, item := range result {
		index[item.ID] = i
	}
	for _, item := range incoming {
		if i, found := index[item.ID]; !found || item.Version >= result[i].Version {
			if found {
				result[i] = item
			} else {
				index[item.ID] = len(result)
				result = append(result, item)
			}
		}
	}
	return result
}

func mergeInspections(base, incoming []domain.Inspection) []domain.Inspection {
	result := append([]domain.Inspection(nil), base...)
	for _, item := range incoming {
		found := false
		for i := range result {
			if result[i].ID == item.ID {
				result[i] = item
				found = true
				break
			}
		}
		if !found {
			result = append(result, item)
		}
	}
	return result
}

func mergeRepairs(base, incoming []domain.RepairRecord) []domain.RepairRecord {
	result := append([]domain.RepairRecord(nil), base...)
	for _, item := range incoming {
		found := false
		for i := range result {
			if result[i].ID == item.ID {
				result[i] = item
				found = true
				break
			}
		}
		if !found {
			result = append(result, item)
		}
	}
	return result
}
