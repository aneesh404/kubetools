package api

import (
	"net/http"

	"github.com/aneeshchawla/kubetools/backend/internal/api/handlers"
	"github.com/aneeshchawla/kubetools/backend/internal/api/middleware"
	"github.com/aneeshchawla/kubetools/backend/internal/services"
)

type Dependencies struct {
	CORSOrigins []string
	Templates   *services.TemplateService
	CRD         *services.CRDService
	YAML        *services.YAMLService
	Manifests   *services.ManifestService
}

func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	crdHandler := handlers.NewCRDHandler(deps.Templates, deps.CRD, deps.YAML, deps.Manifests)

	mux.HandleFunc("/healthz", handlers.Health)
	mux.HandleFunc("/api/v1/health", handlers.Health)
	mux.HandleFunc("/api/v1/crd/templates", crdHandler.Templates)
	mux.HandleFunc("/api/v1/crd/parse", crdHandler.ParseCRD)
	mux.HandleFunc("/api/v1/crd/validate", crdHandler.ValidateCRD)
	mux.HandleFunc("/api/v1/crd/import-url", crdHandler.ImportCRDFromURL)
	mux.HandleFunc("/api/v1/crd/submit", crdHandler.SubmitCRD)
	mux.HandleFunc("/api/v1/crd/generate-yaml", crdHandler.GenerateYAML)
	mux.HandleFunc("/api/v1/manifests", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			crdHandler.ListManifests(w, r)
		case http.MethodPost:
			crdHandler.SaveManifest(w, r)
		default:
			handlers.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only GET and POST are supported")
		}
	})

	return middleware.CORS(deps.CORSOrigins, mux)
}
