package problem

import (
	"fmt"
	"regexp"
	"strings"
)

// ProblemDescription is the description of a problem. It includes the title,
// part 1 description and part 2 description, all of which are in CommonMark
// format.
type ProblemDescription struct {
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
func NewProblemDescription(readme string, input InputGenerator) (ProblemDescription, error) {
	description, err := parseProblemREADME(readme)
	if err != nil {
		return ProblemDescription{}, fmt.Errorf("failed to parse problem README: %w", err)
	}

	description.Input = input
	return description, nil
}

var (
	reTitle = regexp.MustCompile(`(?m)^# (.*)$`)
	rePart1 = regexp.MustCompile(`(?m)^## Part 1$`)
	rePart2 = regexp.MustCompile(`(?m)^## Part 2$`)
)

func parseProblemREADME(md string) (ProblemDescription, error) {
	titleIx := reTitle.FindStringSubmatchIndex(md)
	if titleIx == nil {
		return ProblemDescription{}, fmt.Errorf("failed to find title in README")
	}

	title := md[titleIx[2]:titleIx[3]]
	md = md[titleIx[1]:]
	md = strings.TrimSpace(md)

	// Remove the "Part 1" subtitle.
	part1Idx := rePart1.FindStringIndex(md)
	part1 := strings.TrimSpace(md[:part1Idx[0]]) +
		"\n\n" +
		strings.TrimSpace(md[part1Idx[1]:])

	// Extract the part 2 description.
	part2Idx := rePart2.FindStringIndex(part1)
	part2 := part1[part2Idx[1]:]
	part1 = part1[:part2Idx[0]]

	part1 = strings.TrimSpace(part1)
	part2 = strings.TrimSpace(part2)

	return ProblemDescription{
		Title: title,
		Part1: part1,
		Part2: part2,
	}, nil
}
