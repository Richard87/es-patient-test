package Patient

import (
	"crypto/rand"
	"fmt"
	"github.com/oklog/ulid/v2"
	"time"
)

type Event interface {
	isEvent()
}

func (e Admitted) isEvent()    {}
func (e Transferred) isEvent() {}
func (e Discharged) isEvent()  {}

type ID string
type Name string
type WardNumber int
type Age int

type Admitted struct {
	ID   ID         `json:"id"`
	Name Name       `json:"name"`
	Ward WardNumber `json:"ward"`
	Age  Age        `json:"age"`
}

type Transferred struct {
	ID            ID         `json:"id"`
	NewWardNumber WardNumber `json:"new_ward"`
}

type Discharged struct {
	ID ID `json:"id"`
}

type Patient struct {
	id         ID
	ward       WardNumber
	name       Name
	age        Age
	discharged bool

	changes []Event
	version int
}

func NewFromEvents(events []Event) *Patient {
	p := &Patient{}
	for _, e := range events {
		p.On(e, false)
	}
	return p
}

func (p Patient) Ward() WardNumber {
	return p.ward
}
func (p Patient) Name() Name {
	return p.name
}
func (p Patient) Age() Age {
	return p.age
}
func (p Patient) Discharged() bool {
	return p.discharged
}
func (p Patient) ID() ID {
	return p.id
}
func New(name Name, ward WardNumber, age Age) (*Patient, error) {
	p := &Patient{}

	id, err := ulid.New(ulid.Timestamp(time.Now()), rand.Reader)
	if err != nil {
		return nil, err
	}

	p.raise(Admitted{
		ID:   ID(id.String()),
		Name: name,
		Ward: ward,
		Age:  age,
	})

	return p, nil
}

func (p *Patient) Transfer(newWard WardNumber) error {
	if p.discharged {
		return fmt.Errorf("patient is discharged")
	}

	p.raise(Transferred{
		ID:            p.id,
		NewWardNumber: newWard,
	})

	return nil
}

func (p *Patient) Discharge() error {
	if p.discharged {
		return fmt.Errorf("patient is discharged")
	}

	p.raise(Discharged{
		ID: p.id,
	})
	return nil
}

func (p *Patient) On(event Event, new bool) {
	switch e := event.(type) {
	case Admitted:
		p.id = e.ID
		p.name = e.Name
		p.ward = e.Ward
		p.age = e.Age

	case Transferred:
		p.ward = e.NewWardNumber

	case Discharged:
		p.discharged = true
	}

	if !new {
		p.version++
	}
}

func (p Patient) Events() []Event {
	return p.changes
}
func (p Patient) Version() int {
	return p.version
}
func (p *Patient) raise(event Event) {
	p.changes = append(p.changes, event)
	p.On(event, true)
}

func (p *Patient) ClearEvents(newVersion int) {
	p.changes = []Event{}
	p.version = newVersion
}
