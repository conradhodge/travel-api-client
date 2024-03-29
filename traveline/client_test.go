package traveline_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/andreyvit/diff"
	"github.com/conradhodge/travel-api-client/traveline"
	"github.com/google/go-cmp/cmp"
)

func TestBuildServiceRequest(t *testing.T) {
	when, _ := time.Parse(time.RFC3339, "2020-03-30T12:34:56+01:00")
	naptanCode := "123456789"
	client := traveline.NewClient(
		"TravelineAPI999",
		"letmein",
		&http.Client{},
	)

	request, err := client.BuildServiceRequest("ab7c1e9b-d06f-44cc-b190-4d36fb564386", naptanCode, when)

	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	expectedRequest := `<Siri version="1.0" xmlns="http://www.siri.org.uk/"><ServiceRequest>` +
		`<RequestTimestamp>2020-03-30T12:34:56+01:00</RequestTimestamp><RequestorRef>TravelineAPI999</RequestorRef>` +
		`<StopMonitoringRequest><RequestTimestamp>2020-03-30T12:34:56+01:00</RequestTimestamp>` +
		`<MessageIdentifier>ab7c1e9b-d06f-44cc-b190-4d36fb564386</MessageIdentifier>` +
		`<MonitoringRef>123456789</MonitoringRef></StopMonitoringRequest></ServiceRequest></Siri>`

	if request != expectedRequest {
		t.Fatalf("Request not as expected (~~want ++got):\n%s", diff.CharacterDiff(expectedRequest, request))
	}
}

func TestParseServiceDelivery(t *testing.T) {
	tests := []struct {
		name                  string
		response              string
		expectedDepartureInfo *traveline.MonitoredVehicleJourney
		expectedError         error
	}{
		{
			name: "Response has aimed and expected departure time",
			response: `<Siri xmlns="http://www.siri.org.uk/" version="1.0">
				<ServiceDelivery>
					<ResponseTimestamp>2020-03-30T00:26:39.911+01:00</ResponseTimestamp>
					<StopMonitoringDelivery version="1.0">
						<ResponseTimestamp>2020-03-30T00:26:39.911+01:00</ResponseTimestamp>
						<RequestMessageRef>64ed3eb6-6d84-4f79-ab57-deef38b06431</RequestMessageRef>
						<MonitoredStopVisit>
							<RecordedAtTime>2014-07-01T15:09:20.889+01:00</RecordedAtTime>
							<MonitoringRef>020035811</MonitoringRef>
							<MonitoredVehicleJourney>
								<FramedVehicleJourneyRef>
									<DataFrameRef></DataFrameRef>
									<DatedVehicleJourneyRef></DatedVehicleJourneyRef>
								</FramedVehicleJourneyRef>
								<VehicleMode>bus</VehicleMode>
								<PublishedLineName>42</PublishedLineName>
								<DirectionName>Toddington, The Green</DirectionName>
								<OperatorRef>153</OperatorRef>
								<MonitoredCall>
									<AimedDepartureTime>2014-07-01T15:09:00.000+01:00</AimedDepartureTime>
									<ExpectedDepartureTime>2014-07-01T15:12:00.000+01:00</ExpectedDepartureTime>
								</MonitoredCall>
							</MonitoredVehicleJourney>
						</MonitoredStopVisit>
					</StopMonitoringDelivery>
				</ServiceDelivery>
			</Siri>`,
			expectedDepartureInfo: &traveline.MonitoredVehicleJourney{
				VehicleMode:       "bus",
				PublishedLineName: "42",
				DirectionName:     "Toddington, The Green",
				OperatorRef:       "153",
				MonitoredCall: struct {
					AimedDepartureTime    string "xml:\"AimedDepartureTime\""
					ExpectedDepartureTime string "xml:\"ExpectedDepartureTime\""
				}{
					AimedDepartureTime:    "2014-07-01T15:09:00.000+01:00",
					ExpectedDepartureTime: "2014-07-01T15:12:00.000+01:00",
				},
			},
		},
		{
			name: "Response has aimed but no expected departure time",
			response: `<Siri xmlns="http://www.siri.org.uk/" version="1.0">
				<ServiceDelivery>
					<ResponseTimestamp>2020-03-30T00:26:39.911+01:00</ResponseTimestamp>
					<StopMonitoringDelivery version="1.0">
						<ResponseTimestamp>2020-03-30T00:26:39.911+01:00</ResponseTimestamp>
						<RequestMessageRef>64ed3eb6-6d84-4f79-ab57-deef38b06431</RequestMessageRef>
						<MonitoredStopVisit>
							<RecordedAtTime>2014-07-01T15:09:20.889+01:00</RecordedAtTime>
							<MonitoringRef>020035811</MonitoringRef>
							<MonitoredVehicleJourney>
								<FramedVehicleJourneyRef>
									<DataFrameRef></DataFrameRef>
									<DatedVehicleJourneyRef></DatedVehicleJourneyRef>
								</FramedVehicleJourneyRef>
								<VehicleMode>bus</VehicleMode>
								<PublishedLineName>42</PublishedLineName>
								<DirectionName>Toddington, The Green</DirectionName>
								<OperatorRef>153</OperatorRef>
								<MonitoredCall>
									<AimedDepartureTime>2014-07-01T15:09:00.000+01:00</AimedDepartureTime>
								</MonitoredCall>
							</MonitoredVehicleJourney>
						</MonitoredStopVisit>
					</StopMonitoringDelivery>
				</ServiceDelivery>
			</Siri>`,
			expectedDepartureInfo: &traveline.MonitoredVehicleJourney{
				VehicleMode:       "bus",
				PublishedLineName: "42",
				DirectionName:     "Toddington, The Green",
				OperatorRef:       "153",
				MonitoredCall: struct {
					AimedDepartureTime    string "xml:\"AimedDepartureTime\""
					ExpectedDepartureTime string "xml:\"ExpectedDepartureTime\""
				}{
					AimedDepartureTime: "2014-07-01T15:09:00.000+01:00",
				},
			},
		},
		{
			name: "Invalid response",
			response: `<Siri xmlns="http://www.siri.org.uk/" version="1.0">
				<ServiceDelivery>
			</Siri>`,
			expectedError: errors.New("XML syntax error on line 3: element <ServiceDelivery> closed by </Siri>"),
		},
		{
			name: "No departure times found",
			response: `<Siri xmlns="http://www.siri.org.uk/" version="1.0">
				<ServiceDelivery>
					<ResponseTimestamp>2020-03-30T00:26:39.911+01:00</ResponseTimestamp>
					<StopMonitoringDelivery version="1.0">
						<ResponseTimestamp>2020-03-30T00:26:39.911+01:00</ResponseTimestamp>
						<RequestMessageRef>64ed3eb6-6d84-4f79-ab57-deef38b06431</RequestMessageRef>
					</StopMonitoringDelivery>
				</ServiceDelivery>
			</Siri>`,
			expectedError: traveline.NoTimesFoundError{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := traveline.NewClient(
				"TravelineAPI999",
				"letmein",
				&http.Client{},
			)

			responseInfo, err := client.ParseServiceDelivery(test.response)

			if test.expectedError != nil {
				if err == nil {
					t.Fatalf("Expected error '%s'; got no error", test.expectedError)
				}
				if err.Error() != test.expectedError.Error() {
					t.Fatalf("Expected error '%s'; got '%s'", test.expectedError.Error(), err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error; got '%s'", err)
				}
			}

			if diff := cmp.Diff(test.expectedDepartureInfo, responseInfo); diff != "" {
				t.Errorf("Actual next departure mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestSend(t *testing.T) {
	tests := []struct {
		name             string
		request          string
		response         string
		statusCode       int
		expectedResponse string
		expectedError    error
	}{
		{
			name:             "200 response received from API request",
			request:          "<Siri><ServiceRequest></ServiceRequest></Siri>",
			response:         "<Siri><ServiceDelivery></ServiceDelivery></Siri>",
			statusCode:       http.StatusOK,
			expectedResponse: "<Siri><ServiceDelivery></ServiceDelivery></Siri>",
			expectedError:    nil,
		},
		{
			name:             "401 response received from API request",
			request:          "<Siri><ServiceRequest></ServiceRequest></Siri>",
			response:         "Invalid user credentials",
			statusCode:       http.StatusUnauthorized,
			expectedResponse: "Invalid user credentials",
			expectedError:    errors.New("error status from API: 401"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := NewTestClient(func(req *http.Request) (*http.Response, error) {
				expectedURL := "https://nextbus.mxdata.co.uk/nextbuses/1.0/1"
				if req.URL.String() != expectedURL {
					t.Fatalf("Expected URL: %s, got: %s", expectedURL, req.URL.String())
				}

				expectedContentType := "application/xml"
				if req.Header.Get("Content-type") != expectedContentType {
					t.Fatalf("Expected URL: %s, got: %s", expectedContentType, req.Header.Get("Content-type"))
				}

				return &http.Response{
					StatusCode: test.statusCode,
					// Send response to be tested
					Body: io.NopCloser(bytes.NewBufferString(test.response)),
					// Must be set to non-nil value or it panics
					Header: make(http.Header),
				}, nil
			})

			tlClient := traveline.NewClient(
				"TravelineAPI999",
				"letmein",
				client,
			)
			response, err := tlClient.Send(test.request)

			if test.expectedError != nil {
				if err == nil {
					t.Fatalf("Expected error '%s'; got no error", test.expectedError)
				}
				if err.Error() != test.expectedError.Error() {
					t.Fatalf("Expected error '%s'; got '%s'", test.expectedError.Error(), err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error; got '%s'", err)
				}
			}

			if response != test.expectedResponse {
				t.Fatalf("Unexpected response ~~want ++got):\n%s", diff.CharacterDiff(test.expectedResponse, response))
			}
		})
	}
}

func TestSendDoFails(t *testing.T) {
	client := NewTestClient(func(req *http.Request) (*http.Response, error) {
		// Return an error so that Client.Do(req) fails
		return nil, errors.New("bang")
	})

	tlClient := traveline.NewClient(
		"TravelineAPI999",
		"letmein",
		client,
	)
	_, err := tlClient.Send("")
	if err == nil {
		t.Fatal("Expected error; got no error")
	}
}

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}
func (errReader) Close() error {
	return nil
}

func TestSendBodyReadAllFails(t *testing.T) {
	client := NewTestClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			// Send response to be tested
			Body: errReader{},
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}, nil
	})

	tlClient := traveline.NewClient(
		"TravelineAPI999",
		"letmein",
		client,
	)
	_, err := tlClient.Send("")
	if err == nil {
		t.Fatal("Expected error; got no error")
	}
}
