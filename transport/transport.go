package transport

import (
	"time"
)

// DepartureInfo represents the details for the next departure from a stop
type DepartureInfo struct {
	VehicleMode           string
	LineName              string
	DirectionName         string
	AimedDepartureTime    *time.Time
	ExpectedDepartureTime *time.Time
}

// API represents an API to get travel times for public transport
type API interface {
	// GetNextDepartureTime returns the next departure time at the stop that the NaPTAN code represents
	GetNextDepartureTime(naptanCode string, when time.Time) (*DepartureInfo, error)
}
