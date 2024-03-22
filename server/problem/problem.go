package problem

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"log/slog"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
)

// ModuleConfig is a module that points to a problem.
type ModuleConfig struct {
	Command string `json:"cmd"`
	README  string `json:"readme"`
	ProblemConfig
}

// ProblemConfig contains optional configuration for a problem.
type ProblemConfig struct {
	// PointsPerPart is the number of points awarded for each part of the problem.
	// It overrides the default [PointsPerPart].
	PointsPerPart float64 `json:"points_per_part,omitempty"`
	// ScoringVersion is the version of the scoring function.
	ScoringVersion ScoringVersion `json:"scoring_version,omitempty"`
}

// Problem is a problem that can be solved.
type Problem struct {
	// ID returns the unique ID of the problem.
	// The ID may be in any format, but it must be unique.
	ID string
	// Description returns the description of the problem.
	Description ProblemDescription

	Runner
	ProblemConfig
}

// NewProblem creates a new problem.
func NewProblem(id string, desc ProblemDescription, runner Runner, cfg ProblemConfig) Problem {
	if cfg.PointsPerPart == 0 {
		cfg.PointsPerPart = PointsPerPart
	}
	if cfg.ScoringVersion == 0 {
		cfg.ScoringVersion = latestScoreScalingVersion
	}
	return Problem{
		ID:            id,
		Description:   desc,
		Runner:        runner,
		ProblemConfig: cfg,
	}
}

// NewProblemFromModule creates a new problem from a problem module.
func NewProblemFromModule(module ModuleConfig, logger *slog.Logger) (Problem, error) {
	var z Problem

	description, err := ParseProblemDescriptionFile(module.README)
	if err != nil {
		return z, fmt.Errorf("failed to parse README file at %q: %w", module.README, err)
	}

	runner, err := NewCommandRunner(logger.With("component", "runner"), module.Command)
	if err != nil {
		return z, fmt.Errorf("failed to create command runner %q: %w", module.Command, err)
	}

	return NewProblem(module.README, description, runner, module.ProblemConfig), nil
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
	command := fmt.Sprintf("%s --seed %d %s", p.command, seed, args)
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

type problemCacheKey uint8

const (
	_ problemCacheKey = iota
	inputCacheKey
	part1CacheKey
	part2CacheKey
)

type runnerCacheKey struct {
	id   string
	seed int
	key  problemCacheKey
}

// CachedRunner wraps a Runner and caches the results in a persistent
// database.
type CachedRunner struct {
	cache     *xsync.MapOf[runnerCacheKey, any]
	logger    *slog.Logger
	problemID string
	runner    Runner
}

// NewCachedRunner creates a new cached runner with an existing Badger database.
func NewCachedRunner(logger *slog.Logger, problem Problem) *CachedRunner {
	return &CachedRunner{
		cache:     xsync.NewMapOf[runnerCacheKey, any](),
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

func getCache[T any](
	ctx context.Context,
	c *CachedRunner,
	seed int, pkey problemCacheKey, fn func(context.Context, int) (T, error),
) (T, error) {
	key := runnerCacheKey{c.problemID, seed, pkey}

	logger := c.logger.With(
		"seed", seed,
		"key.id", key.id,
		"key.seed", key.seed)

	v, ok := c.cache.Load(key)
	if ok {
		logger.DebugContext(ctx, "problem cache hit")
		return v.(T), nil
	}

	logger.DebugContext(ctx, "problem cache miss")

	val, err := fn(ctx, seed)
	if err != nil {
		return val, err
	}

	c.cache.LoadOrStore(key, val)
	return val, nil
}

// CacheAllProblems caches all input generators in a persistent database.
// If cacheDBPath is empty, then an in-memory database is used.
func CacheAllProblems(problems []Problem, logger *slog.Logger) {
	for i := range problems {
		problems[i].Runner = NewCachedRunner(logger, problems[i])
	}
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
