package Patient

import (
	"context"
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(dsn string) (*Repository, error) {

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	r := &Repository{
		db: db,
	}

	if err := r.initEventStore(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repository) initEventStore() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			aggregate_id TEXT,
			aggregate_type TEXT,
			event_type TEXT,
			event_data TEXT,
			version INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(aggregate_id,version)
		);
	`)
	return err
}

func (r *Repository) Close() {
	_ = r.db.Close()
}

func (r *Repository) Load(ctx context.Context, id ID) (*Patient, error) {

	queryContext, err := r.db.QueryContext(ctx, `SELECT aggregate_id, event_type, event_data FROM events WHERE aggregate_id = ?`, id)
	if err != nil {
		return nil, err
	}

	var events []Event
	for queryContext.Next() {
		var aggregateID, eventType string
		var eventData []byte
		err = queryContext.Scan(&aggregateID, &eventType, &eventData)
		if err != nil {
			return nil, err
		}

		switch eventType {
		case "Admitted":
			var event Admitted
			err := json.Unmarshal(eventData, &event)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		case "Transferred":
			var event Transferred
			err := json.Unmarshal(eventData, &event)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		case "Discharged":
			var event Discharged
			err := json.Unmarshal(eventData, &event)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		}
	}

	patient := NewFromEvents(events)
	return patient, nil
}

func (r *Repository) Save(ctx context.Context, patient *Patient) error {

	tx, err := r.db.Begin()
	if err != nil {
		panic(err)
	}

	for i, event := range patient.Events() {
		payload, err := json.Marshal(event)
		var eventName string

		switch event.(type) {
		case Admitted:
			eventName = "Admitted"
		case Transferred:
			eventName = "Transferred"
		case Discharged:
			eventName = "Discharged"
		}

		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO events (aggregate_id, aggregate_type, event_type, event_data, version) VALUES (?, ?, ?, ?, ?)`,
			patient.ID(), "Patient", eventName, payload, patient.Version()+i,
		)
		if err != nil {
			panic(err)
		}
	}

	patient.ClearEvents(patient.Version() + len(patient.Events()))

	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	return nil
}
