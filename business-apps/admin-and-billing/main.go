package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/billing"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/database"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/expenses"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/journal"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/meals"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/stats"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/users"
	"github.com/opticSquid/ranjitar_rannaghor/business-apps/admin-and-billing/wallet"
	"github.com/quic-go/quic-go/http3"
	"github.com/rs/cors"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	_ = godotenv.Load()
	// Default to a text logger; switch to JSON in prod below.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	dbPool := database.InitDB()
	defer dbPool.Close()

	mode := os.Getenv("MODE")
	if mode != "prod" && mode != "dev" {
		slog.Error("MODE must be either 'prod' or 'dev'")
		os.Exit(1)
	}
	if mode == "prod" {
		// In production prefer JSON structured logs
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.AllowAll().Handler)
	if mode == "prod" {
		r.Use(altSvcMiddleware)
	}

	r.Route("/api", func(r chi.Router) {
		r.Get("/users", users.GetUsers)
		r.Post("/users", users.CreateUser)
		r.Post("/wallet/recharge", wallet.RechargeWallet)
		r.Get("/daily-entry", journal.GetDailyEntries)
		r.Post("/daily-entry", journal.CreateDailyEntry)
		r.Put("/daily-entry/{id}", journal.UpdateDailyEntry)
		r.Delete("/daily-entry/{id}", journal.DeleteDailyEntry)
		r.Get("/reports/bill", billing.GetBill)
		r.Get("/expenses", expenses.GetExpenses)
		r.Post("/expenses", expenses.CreateExpense)
		r.Put("/expenses/{id}", expenses.UpdateExpense)
		r.Delete("/expenses/{id}", expenses.DeleteExpense)
		r.Get("/dashboard/stats", stats.GetDashboardStats)
		r.Get("/analytics", stats.GetAnalyticsStats)
		r.Post("/meals", meals.CreateMeal)
		r.Get("/meals", meals.GetMeals)
		r.Put("/meals/{id}", meals.UpdateMeal)
		r.Delete("/meals/{id}", meals.DeleteMeal)
	})

	// Serve static files
	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, ".", "dist")
	FileServer(r, "/", http.Dir(filesDir))

	if mode != "prod" {
		// ── Development: plain HTTP (original behaviour preserved) ───────────
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		slog.Info("DEV mode — starting plain HTTP server", "port", port)
		if err := http.ListenAndServe(":"+port, r); err != nil {
			slog.Error("DEV HTTP server failed", "err", err)
			os.Exit(1)
		}
		return
	}
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		slog.Error("DOMAIN must be set in prod mode")
		os.Exit(1)
	}
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache("certs-cache"),
	}

	tlsConfig := &tls.Config{
		GetCertificate: certManager.GetCertificate,
		// ALPN: client and server agree on a protocol during the TLS handshake.
		// Order matters — h3 is preferred, then h2, then http/1.1 as fallback.
		NextProtos: []string{"h3", "h2", "http/1.1"},
		MinVersion: tls.VersionTLS12, // HTTP/2 needs TLS 1.2+; HTTP/3 uses TLS 1.3 internally via QUIC
	}

	tcpServer := &http.Server{
		Addr:      ":443",
		Handler:   r, // your chi router
		TLSConfig: tlsConfig,
		// Timeouts protect against slow-client attacks; always set these in prod.
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	h3Server := &http3.Server{
		Addr:      ":443",
		Handler:   r,         // same chi router
		TLSConfig: tlsConfig, // same cert source — this is what was broken in the original
	}

	go func() {
		slog.Info("Starting ACME/redirect listener on TCP :80...")
		if err := http.ListenAndServe(":80", certManager.HTTPHandler(nil)); err != nil {
			slog.Error("Port 80 listener failed", "err", err)
			os.Exit(1)
		}
	}()

	go func() {
		slog.Info("Starting HTTP/3 (QUIC/UDP) server on :443", "domain", domain)
		if err := h3Server.ListenAndServe(); err != nil {
			slog.Error("HTTP/3 server error", "err", err)
			os.Exit(1)
		}
	}()

	go func() {
		slog.Info("Starting TCP server (HTTP/1.1 + HTTP/2) on :443", "domain", domain)
		if err := tcpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			slog.Error("TCP server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutdown signal received...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Drain in-flight HTTP/1.1 + HTTP/2 requests before exiting
	if err := tcpServer.Shutdown(ctx); err != nil {
		slog.Warn("TCP server shutdown error", "err", err)
	}
	// HTTP/3 close (quic-go does not yet expose graceful request draining)
	if err := h3Server.Close(); err != nil {
		slog.Warn("HTTP/3 server close error", "err", err)
	}

	slog.Info("All servers stopped cleanly.")
}

func altSvcMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Alt-Svc", `h3=":443"; ma=86400`)
		next.ServeHTTP(w, r)
	})
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))

		// If file doesn't exist, serve index.html for SPA
		fpath := filepath.Join(string(root.(http.Dir)), r.URL.Path)
		if _, err := os.Stat(fpath); os.IsNotExist(err) && !strings.HasPrefix(r.URL.Path, "/api") {
			http.ServeFile(w, r, filepath.Join(string(root.(http.Dir)), "index.html"))
			return
		}

		fs.ServeHTTP(w, r)
	})
}
