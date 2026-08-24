package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"go.etcd.io/bbolt"
	"streetlight/domain"
)

var ErrNotFound = errors.New("record not found")

// ErrVersionConflict is returned by optimistic-concurrency updates when the
// record on disk has a different version than the one the caller read. The
// caller should re-read and retry the change against the current version.
var ErrVersionConflict = errors.New("record version conflict")

var bucketNames = [][]byte{
	[]byte("streetlight_poles"), []byte("control_boxes"), []byte("circuit_lines"),
	[]byte("crews"), []byte("work_orders"), []byte("inspections"), []byte("repair_records"),
}

type DB struct {
	db *bbolt.DB
	mu sync.RWMutex
}

func Open(path string) (*DB, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	d, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	store := &DB{db: d}
	if err = store.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range bucketNames {
			if _, e := tx.CreateBucketIfNotExists(name); e != nil {
				return e
			}
		}
		return nil
	}); err != nil {
		_ = d.Close()
		return nil, err
	}
	return store, nil
}

func OpenEphemeral() (*DB, string, error) {
	f, err := os.CreateTemp("", "streetlight-")
	if err != nil {
		return nil, "", err
	}
	path := f.Name()
	if err = f.Close(); err != nil {
		return nil, "", err
	}
	if err = os.Remove(path); err != nil {
		return nil, "", err
	}
	db, err := Open(path)
	if err != nil {
		return nil, "", err
	}
	return db, path, nil
}

func (d *DB) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.db == nil {
		return nil
	}
	err := d.db.Close()
	d.db = nil
	return err
}

func (d *DB) write(bucket, key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.db == nil {
		return fmt.Errorf("database is closed")
	}
	return d.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket([]byte(bucket)).Put([]byte(key), b) })
}

func (d *DB) read(bucket, key string, target any) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.db == nil {
		return fmt.Errorf("database is closed")
	}
	return d.db.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket([]byte(bucket)).Get([]byte(key))
		if value == nil {
			return ErrNotFound
		}
		return json.Unmarshal(value, target)
	})
}

func (d *DB) list(bucket string, target func([]byte) error) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.db == nil {
		return fmt.Errorf("database is closed")
	}
	return d.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(bucket)).ForEach(func(_, value []byte) error { return target(value) })
	})
}

func (d *DB) Delete(bucket, key string) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.db == nil {
		return fmt.Errorf("database is closed")
	}
	return d.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket([]byte(bucket)).Delete([]byte(key)) })
}

func (d *DB) SaveStreetlightPole(value domain.StreetlightPole) error {
	return d.write("streetlight_poles", value.ID, value)
}
func (d *DB) GetStreetlightPole(id string) (domain.StreetlightPole, error) {
	var v domain.StreetlightPole
	return v, d.read("streetlight_poles", id, &v)
}
func (d *DB) ListStreetlightPoles() ([]domain.StreetlightPole, error) {
	out := []domain.StreetlightPole{}
	err := d.list("streetlight_poles", func(b []byte) error {
		var v domain.StreetlightPole
		if e := json.Unmarshal(b, &v); e != nil {
			return e
		}
		out = append(out, v)
		return nil
	})
	return out, err
}

func (d *DB) SaveControlBox(value domain.ControlBox) error {
	return d.write("control_boxes", value.ID, value)
}
func (d *DB) GetControlBox(id string) (domain.ControlBox, error) {
	var v domain.ControlBox
	return v, d.read("control_boxes", id, &v)
}
func (d *DB) ListControlBoxes() ([]domain.ControlBox, error) {
	out := []domain.ControlBox{}
	err := d.list("control_boxes", func(b []byte) error {
		var v domain.ControlBox
		if e := json.Unmarshal(b, &v); e != nil {
			return e
		}
		out = append(out, v)
		return nil
	})
	return out, err
}

func (d *DB) SaveCircuitLine(value domain.CircuitLine) error {
	return d.write("circuit_lines", value.ID, value)
}
func (d *DB) GetCircuitLine(id string) (domain.CircuitLine, error) {
	var v domain.CircuitLine
	return v, d.read("circuit_lines", id, &v)
}
func (d *DB) ListCircuitLines() ([]domain.CircuitLine, error) {
	out := []domain.CircuitLine{}
	err := d.list("circuit_lines", func(b []byte) error {
		var v domain.CircuitLine
		if e := json.Unmarshal(b, &v); e != nil {
			return e
		}
		out = append(out, v)
		return nil
	})
	return out, err
}

func (d *DB) SaveCrew(value domain.Crew) error { return d.write("crews", value.ID, value) }
func (d *DB) GetCrew(id string) (domain.Crew, error) {
	var v domain.Crew
	return v, d.read("crews", id, &v)
}
func (d *DB) ListCrews() ([]domain.Crew, error) {
	out := []domain.Crew{}
	err := d.list("crews", func(b []byte) error {
		var v domain.Crew
		if e := json.Unmarshal(b, &v); e != nil {
			return e
		}
		out = append(out, v)
		return nil
	})
	return out, err
}

func (d *DB) SaveWorkOrder(value domain.WorkOrder) error {
	return d.write("work_orders", value.ID, value)
}
func (d *DB) GetWorkOrder(id string) (domain.WorkOrder, error) {
	var v domain.WorkOrder
	return v, d.read("work_orders", id, &v)
}

// UpdateWorkOrder performs a read-modify-write on a work order inside a single
// bbolt transaction. The mutator receives the order currently on disk and
// returns the new state. The persisted record's Version is checked against the
// mutator's expected version; if they differ (another writer committed between
// the caller's read and this update) the transaction is aborted with
// ErrVersionConflict so the caller can re-read and retry.
//
// Because the read, the version check, and the write run in one bbolt Update,
// no concurrent writer can overwrite the result between the check and the
// write — closing the lost-update window that occurred when dispatch and
// inspection each did a GetWorkOrder followed by a separate SaveWorkOrder.
func (d *DB) UpdateWorkOrder(id string, expectedVersion int, mutator func(domain.WorkOrder) (domain.WorkOrder, error)) (domain.WorkOrder, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.db == nil {
		return domain.WorkOrder{}, fmt.Errorf("database is closed")
	}
	var result domain.WorkOrder
	err := d.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("work_orders"))
		raw := bucket.Get([]byte(id))
		if raw == nil {
			return ErrNotFound
		}
		var current domain.WorkOrder
		if e := json.Unmarshal(raw, &current); e != nil {
			return e
		}
		if current.Version != expectedVersion {
			return ErrVersionConflict
		}
		updated, e := mutator(current)
		if e != nil {
			return e
		}
		encoded, e := json.Marshal(updated)
		if e != nil {
			return e
		}
		if e := bucket.Put([]byte(id), encoded); e != nil {
			return e
		}
		result = updated
		return nil
	})
	return result, err
}
func (d *DB) ListWorkOrders() ([]domain.WorkOrder, error) {
	out := []domain.WorkOrder{}
	err := d.list("work_orders", func(b []byte) error {
		var v domain.WorkOrder
		if e := json.Unmarshal(b, &v); e != nil {
			return e
		}
		out = append(out, v)
		return nil
	})
	return out, err
}

func (d *DB) SaveInspection(value domain.Inspection) error {
	return d.write("inspections", value.ID, value)
}
func (d *DB) GetInspection(id string) (domain.Inspection, error) {
	var v domain.Inspection
	return v, d.read("inspections", id, &v)
}
func (d *DB) ListInspections() ([]domain.Inspection, error) {
	out := []domain.Inspection{}
	err := d.list("inspections", func(b []byte) error {
		var v domain.Inspection
		if e := json.Unmarshal(b, &v); e != nil {
			return e
		}
		out = append(out, v)
		return nil
	})
	return out, err
}

func (d *DB) SaveRepairRecord(value domain.RepairRecord) error {
	return d.write("repair_records", value.ID, value)
}
func (d *DB) GetRepairRecord(id string) (domain.RepairRecord, error) {
	var v domain.RepairRecord
	return v, d.read("repair_records", id, &v)
}
func (d *DB) ListRepairRecords() ([]domain.RepairRecord, error) {
	out := []domain.RepairRecord{}
	err := d.list("repair_records", func(b []byte) error {
		var v domain.RepairRecord
		if e := json.Unmarshal(b, &v); e != nil {
			return e
		}
		out = append(out, v)
		return nil
	})
	return out, err
}
