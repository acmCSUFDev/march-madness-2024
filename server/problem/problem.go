package problem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log/slog"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"dev.acmcsuf.com/march-madness-2024/server/internal/badgerstub"
	"github.com/dgraph-io/badger/v4"

	badgeropts "github.com/dgraph-io/badger/v4/options"
)

// PointsPerPart is the number of points awarded for solving a part of a
// problem.
const PointsPerPart = 100

// Problem is a problem that can be solved.
type Problem struct {
	// ID returns the unique ID of the problem.
	// The ID may be in any format, but it must be unique.
	ID string
	// Description returns the description of the problem.
	Description ProblemDescription

	Runner
}

// NewProblem creates a new problem.
func NewProblem(id string, desc ProblemDescription, runner Runner) Problem {
	return Problem{
		ID:          id,
		Description: desc,
		Runner:      runner,
	}
}

// Runner is a problem runner.
type Runner interface {
	// Input generates the input for the problem.
	Input(ctx context.Context, seed int) (string, error)
	// Part1Solution returns the solution to part 1 of the problem.
	Part1Solution(ctx context.Context, seed int) (int64, error)
	// Part2Solution returns the solution to part 2 of the problem.
	Part2Solution(ctx context.Context, seed int) (int64, error)
}

// CommandRunner implements Runner using a command.
type CommandRunner struct {
	logger  *slog.Logger
	command string
}

// NewCommandRunner creates a new CommandRunner from a command.
// The command must be in the format "command arg1 arg2 ...".
func NewCommandRunner(logger *slog.Logger, cmd string) (*CommandRunner, error) {
	return &CommandRunner{
		logger:  logger.With("runner", "command"),
		command: cmd,
	}, nil
}

// Input implements Problem.
func (p *CommandRunner) Input(ctx context.Context, seed int) (string, error) {
	return p.run(ctx, seed, "")
}

// Part1Solution implements Problem.
func (p *CommandRunner) Part1Solution(ctx context.Context, seed int) (int64, error) {
	s, err := p.run(ctx, seed, "--part1")
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(s, 10, 64)
}

// Part2Solution implements Problem.
func (p *CommandRunner) Part2Solution(ctx context.Context, seed int) (int64, error) {
	s, err := p.run(ctx, seed, "--part2")
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(s, 10, 64)
}

func (p *CommandRunner) run(ctx context.Context, seed int, args string) (string, error) {
	command := p.command + " " + args
	logger := p.logger.With(
		"seed", seed,
		"command", command)

	var buf strings.Builder
	buf.Grow(128)

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = &buf

	start := time.Now()
	err := cmd.Run()
	taken := time.Since(start)

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			logger.ErrorContext(ctx,
				"failed to generate input using command runner",
				"duration", taken,
				"exit_code", exitErr.ExitCode(),
				"stdout", buf.String(),
				"stderr", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to generate input: %w", err)
	}

	logger.DebugContext(ctx,
		"generated input using Python input generator",
		"duration", taken,
		"stdout_len", buf.Len())

	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// CachedRunner wraps a Runner and caches the results in a persistent
// database.
type CachedRunner struct {
	cache     *badger.DB
	logger    *slog.Logger
	problemID string
	runner    Runner
}

// NewCachedRunner creates a new cached runner with an existing Badger database.
func NewCachedRunner(db *badger.DB, logger *slog.Logger, problem Problem) *CachedRunner {
	return &CachedRunner{
		cache:     db,
		logger:    logger.With("runner", "cached"),
		problemID: problem.ID,
		runner:    problem.Runner,
	}
}

// Input implements Problem.
func (c *CachedRunner) Input(ctx context.Context, seed int) (string, error) {
	return getCache(ctx, c, seed, inputCacheKey, c.runner.Input)
}

// Part1Solution implements Problem.
func (c *CachedRunner) Part1Solution(ctx context.Context, seed int) (int64, error) {
	return getCache(ctx, c, seed, part1CacheKey, c.runner.Part1Solution)
}

// Part2Solution implements Problem.
func (c *CachedRunner) Part2Solution(ctx context.Context, seed int) (int64, error) {
	return getCache(ctx, c, seed, part2CacheKey, c.runner.Part2Solution)
}

type problemCacheKey string

const (
	inputCacheKey problemCacheKey = "input"
	part1CacheKey problemCacheKey = "part1"
	part2CacheKey problemCacheKey = "part2"
)

func getCache[T any](
	ctx context.Context,
	c *CachedRunner,
	seed int, key problemCacheKey, fn func(context.Context, int) (T, error),
) (T, error) {

	id := c.problemID
	dbKey := []byte("[v2][" + id + "][" + strconv.Itoa(seed) + "][" + string(key) + "]")

	logger := c.logger.With(
		"seed", seed,
		"key", string(dbKey))

	var val T
	err := c.cache.View(func(tx *badger.Txn) error {
		item, err := tx.Get(dbKey)
		if err != nil {
			return err
		}
		return item.Value(func(bytes []byte) error {
			return json.Unmarshal(bytes, &val)
		})
	})
	if err == nil {
		logger.DebugContext(ctx, "badger cache hit")
		return val, nil
	}

	logger.DebugContext(ctx,
		"badger cache miss",
		"err", err)

	val, err = fn(ctx, seed)
	if err != nil {
		return val, err
	}

	bytes, err := json.Marshal(val)
	if err != nil {
		logger.ErrorContext(ctx,
			"failed to marshal value",
			"err", err)
		return val, fmt.Errorf("failed to marshal value: %w", err)
	}

	if err = c.cache.Update(func(tx *badger.Txn) error {
		return tx.Set(dbKey, bytes)
	}); err != nil {
		logger.ErrorContext(ctx,
			"failed to set value",
			"err", err)
		return val, fmt.Errorf("failed to set value: %w", err)
	}

	return val, nil
}

// CacheAllProblems caches all input generators in a persistent database.
// If cacheDBPath is empty, then an in-memory database is used.
func CacheAllProblems(cacheDBPath string, problems []Problem, logger *slog.Logger) (io.Closer, error) {
	opts := badger.DefaultOptions(cacheDBPath)
	opts.Compression = badgeropts.ZSTD
	opts.ZSTDCompressionLevel = 1
	opts.Logger = badgerstub.New(logger)
	opts.InMemory = cacheDBPath == ""

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	for i := range problems {
		problems[i].Runner = NewCachedRunner(db, logger, problems[i])
	}

	return db, nil
}

// StringToSeed converts a string to a seed.
// It ensures that the seed is small enough that it is reasonable enough to
// cache the input.
func StringToSeed(str string) int {
	// m controls the maximum seed value. The lower the value, the more likely
	// it is that the input will "collide" with another input, meaning that
	// it is cached.
	const m = 64

	hasher := crc32.NewIEEE()
	hasher.Write([]byte(str))
	h := hasher.Sum32()
	s := math.Round(float64(h) / (math.MaxUint32 / m))

	return int(s)
}
