package main

import (
	"api/internal/data"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn               string
		MaxOpenConnection int
		MaxIdleConns      int
		MaxIdleTime       time.Duration
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8000, "api server port number")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("API_DB_DSN"), "PostgresSQL DSN")
	flag.IntVar(&cfg.db.MaxIdleConns, "db-max-idle-conns", 25, "PostgresSQL maximum idle connections")
	flag.IntVar(&cfg.db.MaxOpenConnection, "db-max-open-conns", 25, "PostgresSQL maximum open connections")
	flag.DurationVar(&cfg.db.MaxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgresSQL maximum idle time for connection")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	logger.Info("Database connection established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModel(db),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.MaxOpenConnection)
	db.SetMaxIdleConns(cfg.db.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.MaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ping the database connection pool for an active connection, if the connection couldn't be established after 5 seconds then close the connection
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
