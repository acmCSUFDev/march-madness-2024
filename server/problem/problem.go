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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"libdb.so/february-frenzy/server/internal/badgerstub"

	badgeropts "github.com/dgraph-io/badger/v4/options"
)

// PointsPerPart is the number of points awarded for solving a part of a
// problem.
const PointsPerPart = 100

// Problem describes a problem that can be solved.
type Problem interface {
	// ID returns the unique ID of the problem.
	// The ID may be in any format, but it must be unique.
	ID() string
	// Description returns the description of the problem.
	Description() ProblemDescription
	// Input generates the input for the problem.
	Input(ctx context.Context, seed int) (string, error)
	// Part1Solution returns the solution to part 1 of the problem.
	Part1Solution(ctx context.Context, seed int) (int64, error)
	// Part2Solution returns the solution to part 2 of the problem.
	Part2Solution(ctx context.Context, seed int) (int64, error)
}

// PythonProblem is a helper struct for Python input generators.
// It invokes the Python script according to lib/problem_utils.py.
type PythonProblem struct {
	logger      *slog.Logger
	description ProblemDescription
	module      string
	path        string
	pwd         string
}

// NewPythonProblem creates a new Python input generator.
// It assumes that the Python script is executed as a module named `problem`.
func NewPythonProblem(pwd, path string, logger *slog.Logger) (*PythonProblem, error) {
	description, err := ParseProblemDescriptionDirectory(pwd, path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse problem description: %w", err)
	}

	modulePath := strings.ReplaceAll(filepath.Clean(path), "/", ".") + ".problem"
	return &PythonProblem{
		logger:      logger,
		description: description,
		module:      modulePath,
		path:        path,
		pwd:         pwd,
	}, nil
}

// ID implements Problem.
func (p *PythonProblem) ID() string {
	return fmt.Sprintf("python:%s:%s", p.pwd, p.module)
}

// Description implements Problem.
func (p *PythonProblem) Description() ProblemDescription {
	return p.description
}

// Input implements Problem.
func (p *PythonProblem) Input(ctx context.Context, seed int) (string, error) {
	return p.run(ctx, seed)
}

// Part1Solution implements Problem.
func (p *PythonProblem) Part1Solution(ctx context.Context, seed int) (int64, error) {
	s, err := p.run(ctx, seed, "--part1")
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(s, 10, 64)
}

// Part2Solution implements Problem.
func (p *PythonProblem) Part2Solution(ctx context.Context, seed int) (int64, error) {
	s, err := p.run(ctx, seed, "--part2")
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(s, 10, 64)
}

func (p *PythonProblem) run(ctx context.Context, seed int, args ...string) (string, error) {
	args = append([]string{"-m", p.module, "--seed", strconv.Itoa(seed)}, args...)

	logger := p.logger.With(
		"problem_id", p.ID(),
		"seed", seed,
		"args", args)

	var buf strings.Builder
	buf.Grow(128)

	cmd := exec.CommandContext(ctx, "python3", args...)
	cmd.Dir = p.pwd
	cmd.Stdout = &buf

	start := time.Now()
	err := cmd.Run()
	taken := time.Since(start)

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			logger.DebugContext(ctx,
				"failed to generate input using Python input generator",
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

// CachedProblem wraps a Problem and caches the results in a persistent
// database.
type CachedProblem struct {
	cache   *badger.DB
	logger  *slog.Logger
	problem Problem
}

// NewCachedInputGenerator creates a new cached input generator from an existing
// database.
func NewCachedInputGenerator(db *badger.DB, problem Problem, logger *slog.Logger) *CachedProblem {
	return &CachedProblem{
		cache:   db,
		problem: problem,
		logger:  logger.With("problem_id", problem.ID()),
	}
}

// ID implements Problem.
func (c *CachedProblem) ID() string {
	return c.problem.ID()
}

// Description implements Problem.
func (c *CachedProblem) Description() ProblemDescription {
	return c.problem.Description()
}

// Input implements Problem.
func (c *CachedProblem) Input(ctx context.Context, seed int) (string, error) {
	return getCache(ctx, c, seed, inputCacheKey, c.problem.Input)
}

// Part1Solution implements Problem.
func (c *CachedProblem) Part1Solution(ctx context.Context, seed int) (int64, error) {
	return getCache(ctx, c, seed, part1CacheKey, c.problem.Part1Solution)
}

// Part2Solution implements Problem.
func (c *CachedProblem) Part2Solution(ctx context.Context, seed int) (int64, error) {
	return getCache(ctx, c, seed, part2CacheKey, c.problem.Part2Solution)
}

type problemCacheKey string

const (
	inputCacheKey problemCacheKey = "input"
	part1CacheKey problemCacheKey = "part1"
	part2CacheKey problemCacheKey = "part2"
)

func getCache[T any](
	ctx context.Context,
	c *CachedProblem,
	seed int, key problemCacheKey, fn func(context.Context, int) (T, error),
) (T, error) {

	id := c.problem.ID()
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
		input := NewCachedInputGenerator(db, problems[i], logger)
		problems[i] = input
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
