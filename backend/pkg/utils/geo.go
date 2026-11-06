package utils

import (
	"math"
)

// CalculateDistance uses the Haversine formula to calculate the distance between
// two points on the Earth's surface given their coordinates in decimal degrees.
// Returns the distance in meters.
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusMeters = 6371000

	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}
