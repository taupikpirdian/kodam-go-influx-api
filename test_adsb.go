package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// Test ADS-B data structures
type ADSBData struct {
	Timestamp string        `json:"ts"`
	Count     int           `json:"count"`
	Center    struct {
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lon"`
		RadiusM   int     `json:"radius_m"`
	} `json:"center"`
	Flights []ADSBFlight `json:"flights"`
}

type ADSBFlight struct {
	ID                   string  `json:"id"`
	Callsign             string  `json:"callsign"`
	ICAO24Bit            *string `json:"icao24bit"`
	Registration         string  `json:"registration"`
	Position             struct {
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lon"`
	} `json:"position"`
	AltitudeFt          int     `json:"altitude_ft"`
	HeadingDeg          int     `json:"heading_deg"`
	GroundSpeedKts      int     `json:"groundSpeed_kts"`
	VerticalSpeedFpm    int     `json:"verticalSpeed_fpm"`
	UpdatedAt           string  `json:"updatedAt"`
	Country             string  `json:"country"`
	AircraftTypeICAO    string  `json:"aircraftTypeICAO"`
	AircraftModelText   *string `json:"aircraftModelText"`
	AircraftTypeLabelDB *string `json:"aircraftTypeLabelDB"`
	Images              []ADSBImage `json:"images"`
	OriginAirport       *ADSBAirport `json:"originAirport"`
	DestinationAirport  *ADSBAirport `json:"destinationAirport"`
	IsOnGround          bool    `json:"isOnGround"`
	FR24Link            string  `json:"fr24_link"`
	DistanceKm          float64 `json:"distance_km"`
	Source              string  `json:"source"`
}

type ADSBImage struct {
	Src       string `json:"src"`
	Link      string `json:"link"`
	Copyright string `json:"copyright"`
	Source    string `json:"source"`
}

type ADSBAirport struct {
	IATA      string  `json:"iata"`
	ICAO      string  `json:"icao"`
	Name      string  `json:"name"`
	ShortName *string `json:"shortName"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Position  struct {
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lon"`
	} `json:"position"`
}

func main() {
	// Test sample JSON data (similar to what's in the specification)
	jsonData := `{
		"ts": "2025-11-19T06:11:49.593Z",
		"count": 77,
		"center": {
			"lat": -6.2088,
			"lon": 106.8456,
			"radius_m": 400000
		},
		"flights": [
			{
				"id": "3d2a466b",
				"callsign": "GIA186",
				"icao24bit": null,
				"registration": "PK-GFH",
				"position": {
					"lat": -3.51,
					"lon": 103.986
				},
				"altitude_ft": 34000,
				"heading_deg": 319,
				"groundSpeed_kts": 472,
				"verticalSpeed_fpm": 64,
				"updatedAt": "2025-11-19T06:11:49.591Z",
				"country": "Indonesia",
				"aircraftTypeICAO": "B738",
				"aircraftModelText": null,
				"aircraftTypeLabelDB": "Fixed-wing jet airliner",
				"images": null,
				"originAirport": null,
				"destinationAirport": null,
				"isOnGround": false,
				"fr24_link": "https://www.flightradar24.com/GIA186",
				"distance_km": 436.369,
				"source": "fr24"
			}
		]
	}`

	var ad ADSBData
	if err := json.Unmarshal([]byte(jsonData), &ad); err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	fmt.Printf("Successfully parsed ADSB data:\n")
	fmt.Printf("Timestamp: %s\n", ad.Timestamp)
	fmt.Printf("Count: %d\n", ad.Count)
	fmt.Printf("Center: Lat=%.6f, Lon=%.6f, Radius=%d\n", ad.Center.Latitude, ad.Center.Longitude, ad.Center.RadiusM)
	fmt.Printf("Flights: %d\n", len(ad.Flights))

	if len(ad.Flights) > 0 {
		flight := ad.Flights[0]
		fmt.Printf("First flight:\n")
		fmt.Printf("  Callsign: %s\n", flight.Callsign)
		fmt.Printf("  Registration: %s\n", flight.Registration)
		fmt.Printf("  Altitude: %d ft\n", flight.AltitudeFt)
		fmt.Printf("  Speed: %d kts\n", flight.GroundSpeedKts)
		fmt.Printf("  Position: Lat=%.6f, Lon=%.6f\n", flight.Position.Latitude, flight.Position.Longitude)
	}

	fmt.Println("\nTest completed successfully!")
}