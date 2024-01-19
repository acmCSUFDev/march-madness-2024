package problem

import "time"

// ProblemSet is a set of problems. It is a collection of problems that
// can be solved in any order. It supports timed releases of problems.
type ProblemSet struct {
	problems []Problem
	schedule *ProblemReleaseSchedule
	now      func() time.Time
}

// ProblemReleaseSchedule is the schedule for releasing problems.
type ProblemReleaseSchedule struct {
	// StartReleaseAt is the time at which the first problem is released.
	StartReleaseAt time.Time
	// ReleaseEvery is the duration between releases.
	ReleaseEvery time.Duration
}

// NewProblemSet creates a new problem set.
func NewProblemSet(problems []Problem) *ProblemSet {
	return &ProblemSet{
		problems: problems,
		now:      time.Now,
	}
}

// NewProblemSetWithSchedule creates a new problem set with a release schedule.
func NewProblemSetWithSchedule(problems []Problem, schedule *ProblemReleaseSchedule) *ProblemSet {
	return &ProblemSet{
		problems: problems,
		schedule: schedule,
		now:      time.Now,
	}
}

// Problems returns all available problems in the set.
func (p *ProblemSet) Problems() []Problem {
	return p.problems[:p.AvailableProblems()]
}

// Problem returns the problem at the given index. If the index is accessing a
// problem that is not available yet, it returns nil.
func (p *ProblemSet) Problem(i int) *Problem {
	n := p.AvailableProblems()
	if i < 0 || i >= n {
		return nil
	}
	return &p.problems[i]
}

// TotalProblems returns the total number of problems in the set.
func (p *ProblemSet) TotalProblems() int {
	return len(p.problems)
}

// AvailableProblems returns the number of problems that are available to be
// solved.
func (p *ProblemSet) AvailableProblems() int {
	if p.schedule == nil {
		return p.TotalProblems()
	}

	now := p.now()
	delta := now.Sub(p.schedule.StartReleaseAt)
	if delta < 0 {
		// The first problem is not released yet.
		return 0
	}
	n := int(delta/p.schedule.ReleaseEvery) + 1
	if n >= len(p.problems) {
		// All problems are released.
		return p.TotalProblems()
	}
	return n
}

// NextReleaseTime returns the time at which the next problem will be released.
// If all problems are released, it returns the zero time.
func (p *ProblemSet) NextReleaseTime() time.Time {
	if p.schedule == nil {
		return time.Time{}
	}

	n := p.AvailableProblems()
	if n == p.TotalProblems() {
		return time.Time{}
	}

	return p.schedule.StartReleaseAt.Add(time.Duration(n) * p.schedule.ReleaseEvery)
}

// TimeUntilNextRelease returns the duration until the next problem is released.
// If all problems are released, it returns the zero duration.
func (p *ProblemSet) TimeUntilNextRelease() time.Duration {
	next := p.NextReleaseTime()
	if next.IsZero() {
		return 0
	}
	return next.Sub(p.now())
}
