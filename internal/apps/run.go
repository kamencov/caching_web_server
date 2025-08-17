package apps

import (
	"caching_web_server/internal/apps/config"
	handlerAuth "caching_web_server/internal/handler/auth"
	"caching_web_server/internal/handler/docs/delete"
	"caching_web_server/internal/handler/docs/get"
	"caching_web_server/internal/handler/docs/post"
	"caching_web_server/internal/middleware"
	serviceAuth "caching_web_server/internal/service/auth"
	"caching_web_server/internal/service/docs"
	"caching_web_server/internal/storage/pq"
	"caching_web_server/internal/storage/s3"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Run struct {
}

func NewRun() *Run {
	return &Run{}
}

// Run - запуск приложения
func (r *Run) Run() error {
	cfg := config.New()
	err := cfg.Parse()
	if err != nil {
		return err
	}

	// инициализация логгера
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	// инициализация репозитория
	repoPsql, err := pq.NewStorage(log)
	if err != nil {
		log.Error("Run", "failed to init repoPsql", err)
		return err
	}

	// инициализация Minio
	repoMinio, err := s3.NewMinioStorage(log)
	if err != nil {
		log.Error("run_app", "failed to init minio", err)
		return err
	}

	// инициализация сервиса
	service := serviceAuth.NewService(repoPsql, log, cfg.TokenSalt)
	serviceDocs := docs.NewService(repoPsql, repoMinio, log)

	// инициализация middleware
	middlewareAuth := middleware.NewMiddleware(service, log)

	// регистрация ручек
	handler := handlerAuth.NewHandler(service, log, cfg.AdminToken)
	handlerPostDocs := post.NewHandler(serviceDocs, log, cfg.MaxSizFile)
	handlerGetDocs := get.NewHandler(serviceDocs, log)
	handlerDeleteDocs := delete.NewHandler(serviceDocs, log)

	// запуск сервера
	mux := http.NewServeMux()
	mux.HandleFunc("/api/register", handler.Register)
	mux.HandleFunc("/api/auth", handler.Auth)
	mux.HandleFunc("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			middlewareAuth.Authorize(handlerPostDocs.SaveDocument)(w, r)
		case http.MethodGet:
			middlewareAuth.Authorize(handlerGetDocs.GetDocuments)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/docs/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			middlewareAuth.Authorize(handlerDeleteDocs.DeleteData)(w, r)
		case http.MethodGet:
			middlewareAuth.Authorize(handlerGetDocs.GetDocument)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/auth/{token}", middlewareAuth.Authorize(handler.Logout))

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("Server started", "addr", cfg.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Error("Run", "server error", err)
		}
	}()

	<-ctx.Done()

	log.Info("Shutting down server...")

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Run", "failed to shutdown server", err)
		return err
	}

	log.Info("Shutdown complete")
	return nil
}
