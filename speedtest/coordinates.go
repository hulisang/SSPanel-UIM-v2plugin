package speedtest

import "math"

type Coordinates struct {
	Latitude  float32 `xml:"lat,attr"`
	Longitude float32 `xml:"lon,attr"`
}

const radius = 6371 // km

func radians32(degrees float32) float64 {
	return radians(float64(degrees))
}

func radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func (org Coordinates) DistanceTo(dest Coordinates) float64 {
	dlat := radians32(dest.Latitude - org.Latitude)
	dlon := radians32(dest.Longitude - org.Longitude)
	a := (math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(radians32(org.Latitude))*
			math.Cos(radians32(dest.Latitude))*math.Sin(dlon/2)*
			math.Sin(dlon/2))
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := radius * c

	return d
}
