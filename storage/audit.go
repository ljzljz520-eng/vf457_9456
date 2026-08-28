package storage

import (
	"fmt"
	"streetlight/domain"
)

type Inventory struct {
	Poles        []domain.StreetlightPole
	ControlBoxes []domain.ControlBox
	Circuits     []domain.CircuitLine
	Crews        []domain.Crew
	Orders       []domain.WorkOrder
}

func (d *DB) Inventory() (Inventory, error) {
	poles, err := d.ListStreetlightPoles()
	if err != nil {
		return Inventory{}, err
	}
	boxes, err := d.ListControlBoxes()
	if err != nil {
		return Inventory{}, err
	}
	circuits, err := d.ListCircuitLines()
	if err != nil {
		return Inventory{}, err
	}
	crews, err := d.ListCrews()
	if err != nil {
		return Inventory{}, err
	}
	orders, err := d.ListWorkOrders()
	if err != nil {
		return Inventory{}, err
	}
	return Inventory{Poles: poles, ControlBoxes: boxes, Circuits: circuits, Crews: crews, Orders: orders}, nil
}

func (d *DB) ValidateReferences() error {
	inv, err := d.Inventory()
	if err != nil {
		return err
	}
	poles := map[string]bool{}
	circuits := map[string]bool{}
	crews := map[string]bool{}
	for _, p := range inv.Poles {
		poles[p.ID] = true
	}
	for _, c := range inv.Circuits {
		circuits[c.ID] = true
	}
	for _, c := range inv.Crews {
		crews[c.ID] = true
	}
	for _, b := range inv.ControlBoxes {
		if !poles[b.PoleID] || !circuits[b.CircuitID] {
			return fmt.Errorf("control box %s has missing reference", b.ID)
		}
	}
	for _, o := range inv.Orders {
		if !poles[o.PoleID] || (o.CircuitID != "" && !circuits[o.CircuitID]) {
			return fmt.Errorf("work order %s has missing reference", o.ID)
		}
		if o.AssignedCrew != "" && !crews[o.AssignedCrew] {
			return fmt.Errorf("work order %s has missing crew", o.ID)
		}
	}
	return nil
}
