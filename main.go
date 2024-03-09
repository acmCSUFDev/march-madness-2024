package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/httplog/v2"
	"github.com/lmittmann/tint"
	"github.com/spf13/pflag"
	"libdb.so/february-frenzy/internal/config"
	"libdb.so/february-frenzy/server"
	"libdb.so/february-frenzy/server/db"
	"libdb.so/february-frenzy/server/problem"
	"libdb.so/hserve"
)

var (
	configPath = "config.json"
	verbose    = false
)

func main() {
	pflag.StringVarP(&configPath, "config", "c", configPath, "path to config file")
	pflag.BoolVarP(&verbose, "verbose", "v", verbose, "enable verbose logging")
	pflag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatalln(err)
	}
}

func run(ctx context.Context) error {
	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}

	logOutput := tint.NewHandler(os.Stderr, &tint.Options{
		Level: logLevel,
	})

	logger := slog.New(logOutput)
	slog.SetDefault(logger)

	config, err := config.ParseFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	frontendDir := os.DirFS(config.Paths.Frontend)
	logger.Debug("using frontend dir", "path", config.Paths.Frontend)

	database, err := db.Open(config.Paths.Database)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	secretKey, err := ensureSecretKey(config.Paths.SecretKey)
	if err != nil {
		return fmt.Errorf("failed to ensure secret key exists: %w", err)
	}

	problems := make([]problem.Problem, len(config.Problems.Paths))
	for i, path := range config.Problems.Paths {
		p, err := problem.NewPythonProblem(config.Problems.PWD, path, logger.With("component", "problem"))
		if err != nil {
			return fmt.Errorf("failed to parse problem %q: %w", path, err)
		}
		problems[i] = p
	}

	db, err := problem.CacheAllProblems(config.Paths.ProblemsCache, problems, logger.With("component", "problem_cache"))
	if err != nil {
		return fmt.Errorf("failed to wrap problems with input cache: %w", err)
	}
	defer db.Close()

	problemset := problem.NewProblemSetWithSchedule(problems, &problem.ProblemReleaseSchedule{
		StartReleaseAt: config.Problems.Schedule.Start,
		ReleaseEvery:   config.Problems.Schedule.Every.Duration(),
	})

	server := server.New(server.ServerConfig{
		FrontendDir:          frontendDir,
		SecretKey:            secretKey,
		Problems:             problemset,
		Database:             database,
		Logger:               logger.With("component", "http"),
		HackathonConfig:      config.Hackathon,
		OpenRegistrationTime: config.OpenRegistrationTime,
	})

	handler := http.Handler(server)
	if verbose {
		httpLogger := &httplog.Logger{
			Logger:  logger,
			Options: httplog.Options{LogLevel: slog.LevelDebug},
		}
		middleware := httplog.Handler(httpLogger)
		handler = middleware(handler)
	}

	logger.Info("starting server", "addr", config.HTTPAddress)
	return hserve.ListenAndServe(ctx, config.HTTPAddress, handler)
}

func ensureSecretKey(path string) (server.SecretKey, error) {
	if f, err := os.ReadFile(path); err == nil {
		key, err := server.ParseSecretKey(f)
		if err != nil {
			return server.SecretKey{}, fmt.Errorf("failed to parse secret key file: %w", err)
		}
		return key, nil
	} else if !os.IsNotExist(err) {
		return server.SecretKey{}, fmt.Errorf("failed to open secret key file: %w", err)
	}

	key := server.NewSecretKey()
	if err := os.WriteFile(path, key.ExportBytes(), 0600); err != nil {
		return server.SecretKey{}, fmt.Errorf("failed to write secret key file: %w", err)
	}

	return key, nil
}
