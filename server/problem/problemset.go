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

// Schedule returns the release schedule of the problem set. If the problem set
// does not have a release schedule, it returns nil.
func (p *ProblemSet) Schedule() *ProblemReleaseSchedule {
	return p.schedule
}

// StartedAt returns the time at which the first problem is released. If the
// problem set does not have a release schedule, it returns the zero time.
func (p *ProblemSet) StartedAt() time.Time {
	if p.schedule == nil {
		return time.Time{}
	}
	return p.schedule.StartReleaseAt
}

// EndingAt returns the time at which the last problem is released. If the
// problem set does not have a release schedule, it returns the zero time.
func (p *ProblemSet) EndingAt() time.Time {
	if p.schedule == nil {
		return time.Time{}
	}
	return p.schedule.StartReleaseAt.Add(time.Duration(len(p.problems)) * p.schedule.ReleaseEvery)
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

// ProblemStartTime calculates the time at which the problem at the given index
// was released. If the problem set does not have a release schedule, it returns
// the zero time.
func (p *ProblemSet) ProblemStartTime(i int) time.Time {
	if p.schedule == nil {
		return time.Time{}
	}
	start := p.schedule.StartReleaseAt
	delta := time.Duration(i) * p.schedule.ReleaseEvery
	return start.Add(delta)
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
