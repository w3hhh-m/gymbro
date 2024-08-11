package points

import (
	"math"
)

func CalculatePoints(maxWeight, maxReps, currentWeight, currentReps, base int) int {
	dbMax := float64(maxWeight) * (1 + 0.0333*float64(maxReps))
	currSet := float64(currentWeight) * (1 + 0.0333*float64(currentReps))

	percentage := currSet / dbMax
	points := percentage * float64(base)

	return int(math.Round(points))
}
