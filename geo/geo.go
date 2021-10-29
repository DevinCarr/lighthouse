package geo

import "math"

const r = 6371 // Earth radius in km (approx)

type Point struct {
	Lat  float64
	Long float64
}

func toRadians(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func haver(theta float64) float64 {
	return 0.5 * (1 - math.Cos(theta))
}

func Haversine(p1 Point, p2 Point) float64 {
	phi1 := toRadians(p1.Lat)
	phi2 := toRadians(p2.Lat)
	lambda1 := toRadians(p1.Long)
	lambda2 := toRadians(p2.Long)

	return 2 * r * math.Asin(math.Sqrt(haver(phi2-phi1)+math.Cos(phi1)*math.Cos(phi2)*haver(lambda2-lambda1)))
}
