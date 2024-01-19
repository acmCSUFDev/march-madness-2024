package problem

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// PointsPerPart is the number of points awarded for solving a part of a
// problem.
const PointsPerPart = 100

// Problem is the description of a problem. It includes the title,
// part 1 description and part 2 description, all of which are in CommonMark
// format.
type Problem struct {
	Title string
	Part1 string
	Part2 string
	Input InputGenerator
}

// NewProblemDescription creates a new problem description.
//
// # Parsing README
//
// The README file is assumed to be in CommonMark format. It is parsed with the
// following assumptions:
//   - The first title (`# Title`) is the problem title.
//   - Everything following the "Part 1" subtitle (`## Part 1`) is the part 1
//     description.
//   - Everything following the "Part 2" subtitle (`## Part 2`) is the part 2
//     description.
func NewProblemDescription(readme string, input InputGenerator) (Problem, error) {
	description, err := parseProblemREADME(readme)
	if err != nil {
		return Problem{}, fmt.Errorf("failed to parse problem README: %w", err)
	}

	description.Input = input
	return description, nil
}

var (
	reTitle = regexp.MustCompile(`(?m)^# (.*)$`)
	rePart1 = regexp.MustCompile(`(?m)^## Part 1$`)
	rePart2 = regexp.MustCompile(`(?m)^## Part 2$`)
)

func parseProblemREADME(md string) (Problem, error) {
	titleIx := reTitle.FindStringSubmatchIndex(md)
	if titleIx == nil {
		return Problem{}, fmt.Errorf("failed to find title in README")
	}

	title := md[titleIx[2]:titleIx[3]]
	md = md[titleIx[1]:]
	md = strings.TrimSpace(md)

	part1 := md
	// Remove the "Part 1" subtitle, if any.
	if part1Idx := rePart1.FindStringIndex(md); part1Idx != nil {
		part1 = strings.TrimSpace(md[:part1Idx[0]]) +
			"\n\n" +
			strings.TrimSpace(md[part1Idx[1]:])
	}

	// Extract the part 2 description.
	part2 := ""
	if part2Idx := rePart2.FindStringIndex(part1); part2Idx != nil {
		part2 = part1[part2Idx[1]:]
		part1 = part1[:part2Idx[0]]
	} else {
		return Problem{}, fmt.Errorf("failed to find part 2 in README")
	}

	part1 = strings.TrimSpace(part1)
	part2 = strings.TrimSpace(part2)

	return Problem{
		Title: title,
		Part1: part1,
		Part2: part2,
	}, nil
}

// ParsePythonProblemDirectory parses a Python problem directory. The directory
// structure is assumed to be:
//   - README.md
//   - problem.py
//
// The working directory must be given, as it determines the environment in
// which the Python script is executed. The README file is assumed to be in
// that directory.
func ParsePythonProblemDirectory(pwd, path string, logger *slog.Logger) (Problem, error) {
	readmeFile, err := os.ReadFile(filepath.Join(pwd, path, "README.md"))
	if err != nil {
		return Problem{}, fmt.Errorf("failed to read README.md: %w", err)
	}

	if _, err = os.Stat(filepath.Join(pwd, path, "problem.py")); err != nil {
		return Problem{}, fmt.Errorf("failed to stat problem.py: %w", err)
	}

	pythonModule := strings.ReplaceAll(filepath.Clean(path), "/", ".") + ".problem"
	pythonInput, err := NewPythonInputGenerator(pwd, pythonModule, logger)
	if err != nil {
		return Problem{}, fmt.Errorf("failed to create Python input generator: %w", err)
	}

	return NewProblemDescription(string(readmeFile), pythonInput)
}

// MustParsePythonProblemDirectory is like ParsePythonProblemDirectory, but
// panics if an error occurs.
func MustParsePythonProblemDirectory(pwd, path string, logger *slog.Logger) Problem {
	desc, err := ParsePythonProblemDirectory(pwd, path, logger)
	if err != nil {
		panic(err)
	}
	return desc
}

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
