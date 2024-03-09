package problem

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestParseProblemREADME(t *testing.T) {
	const input = `# The great otter, or something.

I don't know what to put here. Not like it matters.

## Part 1

**How many riddles does it take to get to the center of a tootsie pop?**

## Part 2

**What did part 1 ask?**
`

	desc, err := parseProblemREADME(input)
	assert.NoError(t, err)
	assert.Equal(t, "The great otter, or something.", desc.Title)
	assert.Equal(t, `I don't know what to put here. Not like it matters.

**How many riddles does it take to get to the center of a tootsie pop?**`, desc.Part1)
	assert.Equal(t, "**What did part 1 ask?**", desc.Part2)
}
