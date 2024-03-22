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

// PointsPerPart is the number of points awarded for solving a part of a
// problem.
const PointsPerPart = 100

// ScalePoints scales the points for a problem's part based on the time the
// problem was started and the time the part was solved.
//
// Optional parameters:
//
// - If maxPoints is 0, it is set to PointsPerPart.
// - If version is 0, the latest scoring function is used.
func ScalePoints(t, startedAt time.Time, maxPoints float64, version ScoringVersion) float64 {
	if maxPoints == 0 {
		maxPoints = PointsPerPart
	}
	return version.fn()(t, startedAt) * maxPoints
}

// ScoringVersion is the version of the scoring function.
type ScoringVersion int

const (
	_ ScoringVersion = iota
	V1ScoreScaling
	V2ScoreScaling

	maxScoreScalingVersion // latest = maxScoreScalingVersion - 1
)

const latestScoreScalingVersion = maxScoreScalingVersion - 1

func (v ScoringVersion) IsValid() bool {
	return 0 < v && v < maxScoreScalingVersion
}

func (v ScoringVersion) fn() scoringFn {
	switch v {
	case 0:
		return latestScoreScalingVersion.fn()
	case V1ScoreScaling:
		return scoreScalingV1
	case V2ScoreScaling:
		return scoreScalingV2
	default:
		panic("invalid scoring version")
	}
}

type scoringFn func(t, startedAt time.Time) float64

var (
	_ scoringFn = scoreScalingV1
	_ scoringFn = scoreScalingV2
)

func scoreScalingV1(t, startedAt time.Time) float64 {
	const maxHour = 24
	// https://www.desmos.com/calculator/22el44ng3r
	f1 := func(x float64) float64 { return (math.Atan(-math.Pi*x+math.Pi/2) / 4) + 0.75 }
	f2 := func(x float64) float64 { return f1(x) + (1 - f1(0)) }
	g := func(x float64) float64 { return clamp(f2(x), 0, 1) }
	x := t.Sub(startedAt).Hours() / maxHour
	return g(x)
}

func scoreScalingV2(t, startedAt time.Time) float64 {
	// https://www.desmos.com/calculator/adpqv3xqzr
	const maxHour = 12
	const intensity = 6.7
	const phase = 1.1
	const m = 3.4
	f1 := func(x float64) float64 { return math.Atan(-m*x+phase) / intensity }
	f2 := func(x float64) float64 { return f1(x) + (1 - f1(0)) }
	g := func(x float64) float64 { return clamp(f2(x), 0, 1) }
	x := t.Sub(startedAt).Hours() / maxHour
	return g(x)
}

func clamp(x, minX, maxX float64) float64 {
	return math.Max(minX, math.Min(maxX, x))
}
