package main

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/http-server/handlers/factory"
	"GYMBRO/internal/http-server/handlers/users/oauth"
	"GYMBRO/internal/http-server/handlers/workouts/scheduler"
	session "GYMBRO/internal/http-server/handlers/workouts/sessions"
	mwlogger "GYMBRO/internal/http-server/middleware/logger"
	"GYMBRO/internal/lib/prettylogger"
	"GYMBRO/internal/storage/postgresql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	// Load configuration
	cfg := config.MustLoad()

	// Setup logger
	log := setupLogger(cfg.Env)
	sessionManager := session.NewSessionManager()

	log.Info("Configuration loaded")
	log.Info("Logger loaded")

	// Initialize database
	db, err := postgresql.New(cfg.StoragePath)
	if err != nil {
		log.Error("Error initializing storage", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	log.Info("Storage loaded")

	// Setup router
	router := setupRouter(cfg, log, db, sessionManager)

	// Start scheduler
	workoutsched := scheduler.NewScheduler(log, db, sessionManager, cfg)
	workoutsched.Start()

	// Start server
	startServer(cfg, router, log)
}

// setupLogger configures and returns a logger based on the environment.
func setupLogger(env string) *slog.Logger {
	switch env {
	case "production":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "local":
		//got it from https://github.com/dusted-go/logging
		prettyHandler := prettylogger.NewHandler(&slog.HandlerOptions{
			Level:       slog.LevelDebug,
			AddSource:   false,
			ReplaceAttr: nil,
		})
		return slog.New(prettyHandler)
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}

// setupRouter configures and returns the router with all necessary routes and middleware.
func setupRouter(cfg *config.Config, log *slog.Logger, db *postgresql.Storage, sm *session.Manager) *chi.Mux {
	handlerFactory := factory.NewConcreteHandlerFactory(log, db, db, cfg, sm)

	userHandlerFactory := handlerFactory.GetUsersHandlerFactory()
	middlewareHandlerFactory := handlerFactory.GetMiddlewaresHandlerFactory()
	workoutHandlerFactory := handlerFactory.GetWorkoutsHandlerFactory()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	oauth.NewOAuth(cfg)

	// Protected routes
	router.Group(func(r chi.Router) {
		r.Use(middlewareHandlerFactory.CreateJWTAuthHandler())

		r.Route("/workouts", func(r chi.Router) {
			r.Post("/start", workoutHandlerFactory.CreateStartHandler())

			r.Group(func(r chi.Router) {
				r.Use(middlewareHandlerFactory.CreateActiveSessionHandler())
				r.Post("/end", workoutHandlerFactory.CreateEndHandler())
				r.Route("/records", func(r chi.Router) {
					r.Post("/add", workoutHandlerFactory.CreateAddHandler())
				})
			})
		})
	})

	router.Route("/users", func(r chi.Router) {
		r.Post("/register", userHandlerFactory.CreateRegisterHandler())
		r.Post("/login", userHandlerFactory.CreateLoginHandler())
		r.Get("/logout", userHandlerFactory.CreateLogoutHandler())

		r.Route("/oauth", func(r chi.Router) {
			r.Get("/{provider}/callback", userHandlerFactory.CreateOAuthCallbackHandler())
			r.Get("/{provider}/logout", userHandlerFactory.CreateLogoutHandler())
			r.Get("/{provider}", userHandlerFactory.CreateOAuthLoginHandler())
		})
	})

	return router
}

// startServer configures and starts the HTTP server.
func startServer(cfg *config.Config, router *chi.Mux, log *slog.Logger) {
	srv := http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		WriteTimeout: cfg.Timeout,
		ReadTimeout:  cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Info("Starting server", slog.String("address", cfg.Address))

	if err := srv.ListenAndServe(); err != nil {
		log.Error("Error starting server", slog.Any("error", err))
	}

	log.Error("Server shutdown", slog.String("address", cfg.Address))
}
