package main

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/http-server/handlers/factory"
	"GYMBRO/internal/http-server/handlers/users/oauth"
	mwlogger "GYMBRO/internal/http-server/middleware/logger"
	"GYMBRO/internal/lib/prettylogger"
	"GYMBRO/internal/services"
	"GYMBRO/internal/storage/postgresql"
	"GYMBRO/internal/storage/redis"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("Configuration loaded")
	log.Info("Logger loaded")

	db, err := postgresql.New(cfg.StoragePath)
	if err != nil {
		log.Error("Error initializing storage", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()
	log.Info("Storage loaded")

	sessionManager, err := redis.New(cfg.RedisPath, cfg.RedisPassword, 0)
	if err != nil {
		log.Error("Error initializing session manager", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("Session manager loaded")

	router := setupRouter(cfg, log, db, sessionManager)

	sessionSched := services.NewSessionScheduler(sessionManager, db, cfg, log)
	sessionSched.Start()

	startServer(cfg, router, log)
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case "production":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "local":
		// got it from https://github.com/dusted-go/logging
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

func setupRouter(cfg *config.Config, log *slog.Logger, db *postgresql.Storage, sm *redis.RedisStorage) *chi.Mux {
	handlerFactory := factory.NewConcreteHandlerFactory(log, db, db, sm, cfg)

	userHandlerFactory := handlerFactory.GetUsersHandlerFactory()
	middlewareHandlerFactory := handlerFactory.GetMiddlewaresHandlerFactory()
	workoutHandlerFactory := handlerFactory.GetWorkoutsHandlerFactory()
	recordHandlerFactory := handlerFactory.GetRecordsHandlerFactory()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	oauth.NewOAuth(cfg)

	router.Group(func(r chi.Router) {
		r.Use(middlewareHandlerFactory.CreateJWTAuthHandler())
		r.Route("/workouts", func(r chi.Router) {
			r.Post("/start", workoutHandlerFactory.CreateStartHandler())
			r.Get("/{workoutID}", workoutHandlerFactory.CreateGetWorkoutHandler())

			r.Group(func(r chi.Router) {
				r.Use(middlewareHandlerFactory.CreateActiveSessionHandler())
				r.Post("/end", workoutHandlerFactory.CreateEndHandler())
				r.Route("/records", func(r chi.Router) {
					r.Post("/add", recordHandlerFactory.CreateAddHandler())
					r.Delete("/{recordID}", recordHandlerFactory.CreateDeleteHandler())
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
