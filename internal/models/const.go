package models

import (
	"errors"
	"time"
)

const (
	AuthorizationToken = "Authorization"
	DurationJwtToken   = time.Hour * 24
	MinEntropyBits     = 50
)

type PVZCity string

const (
	CityMoscow PVZCity = "Москва"
	CitySPB    PVZCity = "Санкт-Петербург"
	CityKazan  PVZCity = "Казань"
)

type TypeProduct string

const (
	TypeElectronic TypeProduct = "электроника"
	TypeClothes    TypeProduct = "одежда"
	TypeShoes      TypeProduct = "обувь"
)

type StatusReception string

const (
	StatusInProgress StatusReception = "in_progress"
	StatusClose      StatusReception = "close"
)

var (
	ErrAuthUser   = errors.New("user is not authorized")
	ErrValidation = errors.New("validation error")
)
