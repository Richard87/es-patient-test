package main

import (
	"context"
	"es-patient-test/Patient"
	"fmt"
	"log"
)

type service struct {
	patientRepo *Patient.Repository
	ctx         context.Context
}

func main() {
	repository, err := Patient.NewRepository("file:locked.sqlite?cache=shared")
	if err != nil {
		_ = fmt.Errorf("error: %v", err)
		return
	}
	defer repository.Close()

	ctx := context.Background()

	s := &service{
		patientRepo: repository,
		ctx:         ctx,
	}

	p, err := Patient.New(
		Patient.Name("Richard Hagen"),
		Patient.WardNumber(1),
		Patient.Age(35),
	)
	_ = p.Transfer(Patient.WardNumber(2))
	_ = p.Transfer(Patient.WardNumber(3))

	err = s.patientRepo.Save(s.ctx, p)
	if err != nil {
		log.Println(err)
	}

	_ = p.Discharge()
	err = p.Transfer(Patient.WardNumber(4))
	if err != nil {
		log.Println(err)
	}

	err = s.patientRepo.Save(s.ctx, p)
	if err != nil {
		log.Println(err)
	}
}
