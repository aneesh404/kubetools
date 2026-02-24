package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/aneeshchawla/kubetools/backend/internal/models"
	"github.com/aneeshchawla/kubetools/backend/internal/services"
)

type CRDHandler struct {
	templates *services.TemplateService
	crd       *services.CRDService
	yaml      *services.YAMLService
	manifests *services.ManifestService
}

func NewCRDHandler(
	templateService *services.TemplateService,
	crdService *services.CRDService,
	yamlService *services.YAMLService,
	manifestService *services.ManifestService,
) *CRDHandler {
	return &CRDHandler{
		templates: templateService,
		crd:       crdService,
		yaml:      yamlService,
		manifests: manifestService,
	}
}

func (h *CRDHandler) Templates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only GET is supported")
		return
	}
	WriteSuccess(w, http.StatusOK, h.templates.List())
}

func (h *CRDHandler) ParseCRD(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST is supported")
		return
	}

	var payload models.ParseCRDRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request payload")
		return
	}

	template, err := h.crd.ParseCRD(payload.Raw)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_CRD", err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, models.ParseCRDResponse{Template: template})
}

func (h *CRDHandler) ValidateCRD(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST is supported")
		return
	}

	var payload models.ValidateCRDRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request payload")
		return
	}

	result := h.crd.ValidateCRD(payload.Raw)
	if !result.Valid {
		WriteSuccess(w, http.StatusOK, result)
		return
	}

	WriteSuccess(w, http.StatusOK, result)
}

func (h *CRDHandler) GenerateYAML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST is supported")
		return
	}

	var payload models.GenerateYAMLRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request payload")
		return
	}

	yamlOutput, err := h.yaml.GenerateYAML(payload.APIVersion, payload.Kind, payload.Fields)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "GENERATION_FAILED", err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, models.GenerateYAMLResponse{YAML: yamlOutput})
}

func (h *CRDHandler) SaveManifest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST is supported")
		return
	}

	var payload models.SaveManifestRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request payload")
		return
	}

	record, err := h.manifests.SaveManifest(r.Context(), payload)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "MANIFEST_SAVE_FAILED", err.Error())
		return
	}

	WriteSuccess(w, http.StatusCreated, record)
}

func (h *CRDHandler) SubmitCRD(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST is supported")
		return
	}

	var payload models.SubmitCRDRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request payload")
		return
	}

	validation := h.crd.ValidateCRD(payload.Raw)
	if !validation.Valid {
		WriteSuccess(w, http.StatusOK, models.SubmitCRDResponse{
			Validation: validation,
		})
		return
	}

	template, err := h.crd.ParseCRD(payload.Raw)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_CRD", err.Error())
		return
	}

	generatedYAML, err := h.yaml.GenerateYAML(template.APIVersion, template.Kind, template.DefaultFields)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "GENERATION_FAILED", err.Error())
		return
	}

	record, err := h.manifests.SaveManifest(r.Context(), models.SaveManifestRequest{
		Title:      fallbackTitle(payload.Title, template.Kind),
		Resource:   template.Kind + " (" + template.APIVersion + ")",
		APIVersion: template.APIVersion,
		Kind:       template.Kind,
		YAML:       generatedYAML,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, "MANIFEST_SAVE_FAILED", err.Error())
		return
	}

	WriteSuccess(w, http.StatusCreated, models.SubmitCRDResponse{
		Template:   template,
		Manifest:   record,
		Validation: validation,
	})
}

func (h *CRDHandler) ImportCRDFromURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST is supported")
		return
	}

	var payload models.ImportCRDURLRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request payload")
		return
	}

	sourceURL, raw, err := h.crd.FetchCRDFromURL(payload.URL)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "CRD_IMPORT_FAILED", err.Error())
		return
	}

	validation := h.crd.ValidateCRD(raw)
	WriteSuccess(w, http.StatusOK, models.ImportCRDURLResponse{
		SourceURL:  sourceURL,
		Raw:        raw,
		Validation: validation,
	})
}

func (h *CRDHandler) ListManifests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only GET is supported")
		return
	}

	query := r.URL.Query().Get("query")
	limit := int64(50)
	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			limit = parsed
		}
	}

	items, err := h.manifests.ListManifests(r.Context(), query, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "MANIFEST_LIST_FAILED", err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, items)
}

func fallbackTitle(title string, kind string) string {
	if strings.TrimSpace(title) != "" {
		return strings.TrimSpace(title)
	}
	if kind != "" {
		return "CRD: " + kind
	}
	return "Submitted CRD"
}
