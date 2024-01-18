package problem

import (
	"context"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/neilotoole/slogt"
)

func TestPythonInputGenerator(t *testing.T) {
	logger := slogt.New(t)

	// We're in ./server/problem.
	// The problem is in ./problems/01.
	gen, err := NewPythonInputGenerator("../../", "problems.01.problem", logger)
	assert.NoError(t, err)

	for i := 0; i < 5; i++ {
		// Ensure deterministic output.
		input, err := gen.GenerateInput(context.Background(), 0)
		assert.NoError(t, err)

		// t.Logf("input: %#v", input)

		expect := ProblemInput{
			Input: "[ OK ] ocserv\n[ OK ] mediawiki\n[STOP] loader\n[ OK ] zerotierone\n[ OK ] fuse\n[ OK ] thefuck",
			Part1: 66,
			Part2: 809,
		}

		if !strings.HasPrefix(input.Input, expect.Input) {
			t.Errorf("test iteration %d returned unexpected output:\n"+
				"want: %q\n"+
				"got:  %q",
				i, expect, input.Input)
		}

		assert.Equal(t, expect.Part1, input.Part1)
		assert.Equal(t, expect.Part2, input.Part2)
	}
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
