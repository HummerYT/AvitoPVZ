package models

import (
	"time"

	"github.com/google/uuid"
)

type PVZ struct {
	ID               string    `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}

type Reception struct {
	ID       uuid.UUID       `json:"id"`
	DateTime time.Time       `json:"dateTime"`
	PvzID    uuid.UUID       `json:"pvzId"`
	Status   StatusReception `json:"status"`
}

type Product struct {
	ID          uuid.UUID   `json:"id"`
	DateTime    time.Time   `json:"dateTime"`
	Type        TypeProduct `json:"type"`
	ReceptionID uuid.UUID   `json:"receptionId"`
}

type ReceptionData struct {
	Reception Reception `json:"reception"`
	Products  []Product `json:"products"`
}

type PVZData struct {
	PVZ        PVZ             `json:"pvz"`
	Receptions []ReceptionData `json:"receptions"`
}

func IsPVZCity(city PVZCity) bool {
	allowedCities := map[PVZCity]bool{
		CityMoscow: true,
		CitySPB:    true,
		CityKazan:  true,
	}

	return allowedCities[city]
}

func IsTypeProduct(product string) bool {
	data := TypeProduct(product)
	allowedTypes := map[TypeProduct]bool{
		TypeElectronic: true,
		TypeClothes:    true,
		TypeShoes:      true,
	}

	return allowedTypes[data]
}
