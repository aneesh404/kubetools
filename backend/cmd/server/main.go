package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aneeshchawla/kubetools/backend/internal/api"
	"github.com/aneeshchawla/kubetools/backend/internal/config"
	"github.com/aneeshchawla/kubetools/backend/internal/services"
)

func main() {
	cfg := config.Load()

	templateService := services.NewTemplateService()
	crdService := services.NewCRDService()
	yamlService := services.NewYAMLService()
	manifestService, err := services.NewManifestService(context.Background(), cfg)
	if err != nil {
		log.Printf("initialize manifest service: %v (falling back to in-memory history)", err)
	}
	if manifestService == nil {
		log.Fatalf("initialize manifest service: no service available")
	}

	router := api.NewRouter(api.Dependencies{
		CORSOrigins: cfg.CORSOrigins,
		Templates:   templateService,
		CRD:         crdService,
		YAML:        yamlService,
		Manifests:   manifestService,
	})

	server := &http.Server{
		Addr:         cfg.Host + ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("backend listening on http://%s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	waitForShutdown(server, manifestService)
}

func waitForShutdown(server *http.Server, manifestService *services.ManifestService) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
		return
	}
	if err := manifestService.Close(ctx); err != nil {
		log.Printf("mongodb disconnect error: %v", err)
	}
	log.Print("server shutdown complete")
}
