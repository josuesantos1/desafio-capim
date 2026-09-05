package clinic

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/josuesantos1/desafio/pkg/problem"
)

func RegisterRoutes(r chi.Router, svc *Service) {
	h := &handler{svc: svc}
	r.Post("/clinics", h.create)
	r.Get("/clinics", h.list)
	r.Get("/clinics/{id}", h.get)
	r.Put("/clinics/{id}", h.update)
	r.Delete("/clinics/{id}", h.delete)
}

type handler struct {
	svc *Service
}

// @Summary      Create a clinic
// @Tags         clinics
// @Accept       json
// @Produce      json
// @Param        clinic  body      CreateInput  true  "Clinic to create"
// @Success      201     {object}  clinicResponse
// @Failure      400     {object}  problem.Details
// @Failure      409     {object}  problem.Details
// @Router       /clinics [post]
func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		problem.Write(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}

	c, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	problem.WriteJSON(w, http.StatusCreated, toClinicResponse(c))
}

// @Summary      Get a clinic by id
// @Tags         clinics
// @Produce      json
// @Param        id   path      string  true  "Clinic ID"
// @Success      200  {object}  clinicResponse
// @Failure      400  {object}  problem.Details
// @Failure      404  {object}  problem.Details
// @Router       /clinics/{id} [get]
func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	c, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	problem.WriteJSON(w, http.StatusOK, toClinicResponse(c))
}

// @Summary      Update a clinic (partial)
// @Tags         clinics
// @Accept       json
// @Produce      json
// @Param        id      path      string       true  "Clinic ID"
// @Param        clinic  body      UpdateInput  true  "Fields to update"
// @Success      200     {object}  clinicResponse
// @Failure      400     {object}  problem.Details
// @Failure      404     {object}  problem.Details
// @Router       /clinics/{id} [put]
func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var in UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		problem.Write(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}

	c, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	problem.WriteJSON(w, http.StatusOK, toClinicResponse(c))
}

// @Summary      Soft-delete a clinic
// @Tags         clinics
// @Param        id   path  string  true  "Clinic ID"
// @Success      204
// @Failure      400  {object}  problem.Details
// @Failure      404  {object}  problem.Details
// @Router       /clinics/{id} [delete]
func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary      List/search clinics (paginated)
// @Tags         clinics
// @Produce      json
// @Param        limit   query     int     false  "Page size (default 20, max 100)"
// @Param        offset  query     int     false  "Offset (default 0)"
// @Param        q       query     string  false  "Free-text search (trade name, legal name, specialties)"
// @Param        city    query     string  false  "Filter by city (substring, case-insensitive)"
// @Success      200     {object}  listResponse
// @Router       /clinics [get]
func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	params := ListParams{
		Limit:  clampLimit(parseQueryInt(r, "limit")),
		Offset: clampOffset(parseQueryInt(r, "offset")),
		Query:  r.URL.Query().Get("q"),
		City:   r.URL.Query().Get("city"),
	}

	result, err := h.svc.List(r.Context(), params)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	problem.WriteJSON(w, http.StatusOK, toListResponse(result, params))
}

func parseQueryInt(r *http.Request, key string) int {
	v, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return 0
	}
	return v
}

func parseID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		problem.Write(w, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return "", false
	}
	return id, true
}

func writeDomainError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		problem.Write(w, http.StatusBadRequest, "VALIDATION_ERROR", "request validation failed", validationErr.Fields)
	case errors.Is(err, ErrDocumentImmutable):
		problem.Write(w, http.StatusBadRequest, "DOCUMENT_IMMUTABLE", "document cannot be changed after creation", nil)
	case errors.Is(err, ErrDocumentExists):
		problem.Write(w, http.StatusConflict, "DOCUMENT_ALREADY_EXISTS", "document already belongs to another clinic", nil)
	case errors.Is(err, ErrNotFound):
		problem.Write(w, http.StatusNotFound, "CLINIC_NOT_FOUND", "clinic not found", nil)
	default:
		slog.Error("unclassified clinic repository error", "error", err)
		problem.Write(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", nil)
	}
}
