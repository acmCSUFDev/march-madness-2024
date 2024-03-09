package problem

import (
	"context"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/neilotoole/slogt"
)

func TestPythonProblem(t *testing.T) {
	logger := slogt.New(t)

	// We're in ./server/problem.
	// The problem is in ./problems/01.
	problem, err := NewPythonProblem("../../", "problems/01", logger)
	assert.NoError(t, err, "cannot create PythonProblem at /problems/01")

	for i := 0; i < 5; i++ {
		// Ensure deterministic output.
		got := getProblemInputAndSolutions(t, problem)
		expect := problemInputAndSolutions{
			Input: "[ OK ] ocserv\n[ OK ] mediawiki\n[STOP] loader\n[ OK ] zerotierone\n[ OK ] fuse\n[ OK ] thefuck",
			Part1: 66,
			Part2: 809,
		}

		if !strings.HasPrefix(got.Input, expect.Input) {
			t.Errorf("test iteration %d returned unexpected output:\n"+
				"want: %q\n"+
				"got:  %q",
				i, expect, got.Input)
		}

		assert.Equal(t, expect.Part1, got.Part1, "part 1 solution mismatch")
		assert.Equal(t, expect.Part2, got.Part2, "part 2 solution mismatch")
	}
}

type problemInputAndSolutions struct {
	Input string
	Part1 int64
	Part2 int64
}

func getProblemInputAndSolutions(t *testing.T, problem Problem) problemInputAndSolutions {
	const seed = 0

	input, err := problem.Input(context.Background(), seed)
	assert.NoError(t, err, "cannot get input")

	part1, err := problem.Part1Solution(context.Background(), seed)
	assert.NoError(t, err, "cannot get part 1 solution")

	part2, err := problem.Part2Solution(context.Background(), seed)
	assert.NoError(t, err, "cannot get part 2 solution")

	return problemInputAndSolutions{input, part1, part2}
}

func TestStringToSeed(t *testing.T) {
	tests := []struct {
		in  string
		out int
	}{
		{"diamondburned", 23},
		{"aaronlieb", 38},
	}

	for _, test := range tests {
		got := StringToSeed(test.in)
		if test.out != got {
			t.Errorf("StringToSeed(%q) = %v != %v", test.in, got, test.out)
		}
	}
}
