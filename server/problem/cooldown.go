package problem

import "time"

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
