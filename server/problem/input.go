package problem

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log/slog"
	"math"
	"os/exec"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v4"
	badgeropts "github.com/dgraph-io/badger/v4/options"
	"libdb.so/february-frenzy/server/internal/badgerstub"
)

// ProblemInput is the expected input and output for a problem.
type ProblemInput struct {
	// Input is the input data for the problem.
	Input string
	// Part1 is the expected output for part 1.
	Part1 int64
	// Part2 is the expected output for part 2.
	Part2 int64
}

// InputGenerator is the interface for a problem input generator.
type InputGenerator interface {
	// GenerateInput generates the input for the problem.
	// Each input is unique to the seed. It is not guaranteed that a different
	// seed will generate a different input. The generator may also choose to
	// cache the input for a seed.
	GenerateInput(ctx context.Context, seed int) (ProblemInput, error)
}

// PythonInputGenerator is a helper struct for Python input generators.
// It invokes the Python script according to lib/problem_utils.py.
type PythonInputGenerator struct {
	logger *slog.Logger
	module string
	pwd    string
}

// NewPythonInputGenerator creates a new Python input generator.
// It assumes that the Python script is executed as a module.
func NewPythonInputGenerator(pwd, modulePath string, logger *slog.Logger) (*PythonInputGenerator, error) {
	return &PythonInputGenerator{
		logger: logger,
		module: modulePath,
		pwd:    pwd,
	}, nil
}

// GenerateInput implements InputGenerator.
func (p *PythonInputGenerator) GenerateInput(ctx context.Context, seed int) (ProblemInput, error) {
	cmd := exec.CommandContext(ctx,
		"python3", "-m", p.module, "--seed", strconv.Itoa(seed), "--json")
	cmd.Dir = p.pwd

	start := time.Now()
	output, err := cmd.CombinedOutput()
	taken := time.Since(start)

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			p.logger.DebugContext(ctx,
				"failed to generate input using Python input generator",
				"seed", seed,
				"module", p.module,
				"pwd", p.pwd,
				"duration", taken,
				"exit_code", exitErr.ExitCode(),
				"stdout", string(output),
				"stderr", string(exitErr.Stderr))
		}
		return ProblemInput{}, fmt.Errorf("failed to generate input: %w", err)
	}

	p.logger.DebugContext(ctx,
		"generated input using Python input generator",
		"seed", seed,
		"module", p.module,
		"pwd", p.pwd,
		"duration", taken)

	var input ProblemInput
	if err := json.Unmarshal(output, &input); err != nil {
		return ProblemInput{}, fmt.Errorf("failed to parse input: %w", err)
	}

	return input, nil
}

// CachedInputGenerator wraps an input generator and caches the results in a
// persistent database.
type CachedInputGenerator struct {
	cache     *badger.DB
	logger    *slog.Logger
	generator InputGenerator
	id        string
}

// NewCachedInputGenerator creates a new cached input generator from an existing
// database.
func NewCachedInputGenerator(db *badger.DB, id string, generator InputGenerator, logger *slog.Logger) *CachedInputGenerator {
	return &CachedInputGenerator{cache: db, generator: generator, id: id, logger: logger}
}

// GenerateInput implements InputGenerator.
func (c *CachedInputGenerator) GenerateInput(ctx context.Context, seed int) (ProblemInput, error) {
	idHash := sha256.Sum256([]byte(c.id))
	idHashStr := base64.StdEncoding.EncodeToString(idHash[:])
	k := []byte("v1-" + idHashStr + "|" + strconv.Itoa(seed))

	var input ProblemInput
	err := c.cache.View(func(tx *badger.Txn) error {
		item, err := tx.Get(k)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &input)
		})
	})
	if err == nil {
		c.logger.DebugContext(ctx,
			"badger cache hit",
			"seed", seed,
			"key", string(k))
		return input, nil
	}

	c.logger.DebugContext(ctx,
		"badger cache miss",
		"seed", seed,
		"key", string(k),
		"err", err)

	input, err = c.generator.GenerateInput(ctx, seed)
	if err != nil {
		return ProblemInput{}, err
	}

	val, err := json.Marshal(input)
	if err != nil {
		return ProblemInput{}, err
	}

	err = c.cache.Update(func(tx *badger.Txn) error {
		return tx.Set(k, val)
	})
	if err != nil {
		return ProblemInput{}, err
	}

	return input, nil
}

// WrapAllProblemsWithInputCache wraps the given problem descriptions with an
// input cache using [CachedInputGenerator].
func WrapAllProblemsWithInputCache(cacheDBPath string, problems []Problem, ids []string, logger *slog.Logger) (io.Closer, error) {
	if len(problems) != len(ids) {
		return nil, fmt.Errorf("length of problems and ids must be equal")
	}

	opts := badger.DefaultOptions(cacheDBPath)
	opts.Compression = badgeropts.ZSTD
	opts.ZSTDCompressionLevel = 1
	opts.Logger = badgerstub.New(logger)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	for i := range problems {
		input := NewCachedInputGenerator(
			db, ids[i], problems[i].Input,
			logger.With("problem_id", ids[i]))
		problems[i].Input = input
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
