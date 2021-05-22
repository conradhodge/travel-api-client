package transport

import (
	"time"

	"github.com/conradhodge/travel-api-client/traveline"
	"github.com/google/uuid"
)

// Traveline is used to make transport requests using the Traveline API
type Traveline struct {
	API traveline.API
}

// NewTraveline returns the implementation of the transport API using the Traveline API
func NewTraveline(api traveline.API) *Traveline {
	return &Traveline{API: api}
}

// GetNextDepartureTime returns the next departure time at the stop that the NaPTAN code represents
func (c *Traveline) GetNextDepartureTime(naptanCode string, when time.Time) (*DepartureInfo, error) {
	request, err := c.API.BuildServiceRequest(uuid.New().String(), naptanCode, when)
	if err != nil {
		return nil, err
	}

	response, err := c.API.Send(request)
	if err != nil {
		return nil, err
	}

	monitoredVehicleJourney, err := c.API.ParseServiceDelivery(response)
	if err != nil {
		return nil, err
	}

	nextDepartureInfo := DepartureInfo{
		LineName:      monitoredVehicleJourney.PublishedLineName,
		VehicleMode:   monitoredVehicleJourney.VehicleMode,
		DirectionName: monitoredVehicleJourney.DirectionName,
	}

	// Convert aimed departure time to time.Time
	aimedDepartureTime, err := convertDepartureTime(monitoredVehicleJourney.MonitoredCall.AimedDepartureTime)
	if err != nil {
		return nil, err
	}
	nextDepartureInfo.AimedDepartureTime = &aimedDepartureTime

	// Convert expected departure time to time.Time
	if len(monitoredVehicleJourney.MonitoredCall.ExpectedDepartureTime) > 0 {
		expectedDepartureTime, err := convertDepartureTime(monitoredVehicleJourney.MonitoredCall.ExpectedDepartureTime)
		if err != nil {
			return nil, err
		}
		nextDepartureInfo.ExpectedDepartureTime = &expectedDepartureTime
	}

	return &nextDepartureInfo, nil
}

func convertDepartureTime(departureTime string) (time.Time, error) {
	convertedDepartureTime, err := time.Parse(time.RFC3339, departureTime)
	if err != nil {
		return time.Time{}, &InvalidTimeFoundError{
			Time:   departureTime,
			Reason: err.Error(),
		}
	}

	return convertedDepartureTime, nil
}
