package main

import (
	"context"
	"crypto/tls"
	"log"
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
	dbPool := database.InitDB()
	defer dbPool.Close()

	mode := os.Getenv("MODE")

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
		log.Printf("DEV mode — starting plain HTTP server on :%s\n", port)
		log.Fatal(http.ListenAndServe(":"+port, r))
		return
	}
	domain := os.Getenv("DOMAIN")
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
		log.Println("Starting ACME/redirect listener on TCP :80...")
		if err := http.ListenAndServe(":80", certManager.HTTPHandler(nil)); err != nil {
			log.Fatalf("Port 80 listener failed: %v", err)
		}
	}()

	go func() {
		log.Printf("Starting HTTP/3 (QUIC/UDP) server on :443 for %s\n", domain)
		if err := h3Server.ListenAndServe(); err != nil {
			log.Fatalf("HTTP/3 server error: %v", err)
		}
	}()

	go func() {
		log.Printf("Starting TCP server (HTTP/1.1 + HTTP/2) on :443 for %s\n", domain)
		if err := tcpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatalf("TCP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Drain in-flight HTTP/1.1 + HTTP/2 requests before exiting
	if err := tcpServer.Shutdown(ctx); err != nil {
		log.Printf("TCP server shutdown error: %v", err)
	}
	// HTTP/3 close (quic-go does not yet expose graceful request draining)
	if err := h3Server.Close(); err != nil {
		log.Printf("HTTP/3 server close error: %v", err)
	}

	log.Println("All servers stopped cleanly.")
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
