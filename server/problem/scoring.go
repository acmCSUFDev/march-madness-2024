package problem

import (
	"math"
	"time"
)

// CalculateCooldownEnd calculates the end of the cooldown period for a problem
// given the total number of attempts, the time of the last submission and the
// current time. If there is no cooldown, the returned timestamp is before the
// current time, which may or may not be zero.
func CalculateCooldownEnd(totalAttempts int, lastSubmitted, now time.Time) time.Time {
	const cooldownThreshold = 2
	const cooldownMultiplier = 2
	const cooldownMax = 5 * time.Minute
	const cooldown = 30 * time.Second

	if totalAttempts < cooldownThreshold {
		return time.Time{}
	}

	n := totalAttempts - cooldownThreshold + 1

	return lastSubmitted.Add(min(
		cooldown*time.Duration(cooldownMultiplier*n),
		cooldownMax))
}

const (
	// PointsPerPart is the number of points awarded for solving a part of a
	// problem.
	PointsPerPart = 100
	// MaxHour is the maximum hour before people get the lowest points.
	MaxHour = 24
)

// ScalePoints scales the points for a problem's part based on the time the
// problem was started and the time the part was solved.
func ScalePoints(t, startedAt time.Time) float64 {
	h := t.Sub(startedAt).Hours()
	return scoreScalingFn(clamp(h/MaxHour, 0, 1)) * PointsPerPart
}

func scoreScalingFn(x float64) float64 {
	// https://www.desmos.com/calculator/22el44ng3r
	f := func(x float64) float64 { return (math.Atan(-math.Pi*x+math.Pi/2) / 4) + 0.75 }
	g := func(x float64) float64 { return f(x) + (1 - f(0)) }
	return g(x)
}

func clamp(x, minX, maxX float64) float64 {
	return math.Max(minX, math.Min(maxX, x))
}
