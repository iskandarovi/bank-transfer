package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"bank-transfer/handlers"
	"bank-transfer/repositories"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := loadEnvFile(".env"); err != nil {
		log.Printf(".env faylini o'qishda ogohlantirish: %v", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL belgilanmagan. .env faylini yarating yoki environment variable o'rnating")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Database ulanishida xato: %v", err)
	}
	defer db.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.Ping(pingCtx); err != nil {
		log.Fatalf("Database ping xatosi: %v", err)
	}
	log.Println("PostgreSQL connected")

	accountRepo := &repositories.AccountRepository{DB: db}
	transferRepo := &repositories.TransferRepository{DB: db}

	accountHandler := &handlers.AccountHandler{Repo: accountRepo}
	transferHandler := &handlers.TransferHandler{Repo: transferRepo}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /accounts", accountHandler.GetAll)
	mux.HandleFunc("GET /accounts/{id}", accountHandler.GetByID)
	mux.HandleFunc("POST /transfer", transferHandler.Transfer)
	mux.HandleFunc("GET /transfers", transferHandler.GetAll)
	mux.HandleFunc("GET /transfers/{id}", transferHandler.GetByID)

	server := &http.Server{
		Addr:              envOrDefault("SERVER_ADDR", ":8080"),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server %s manzilda ishlayapti", server.Addr)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server xatosi: %v", err)
		}
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Serverni to'xtatishda xato: %v", err)
		}
	}
}

func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, "\"'")
		if key != "" && os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
