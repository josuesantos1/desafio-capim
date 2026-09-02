package clinic

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func RegisterRoutes(r chi.Router, svc *Service) {
	h := &handler{svc: svc}
	r.Post("/clinics", h.create)
	r.Get("/clinics/{id}", h.get)
	r.Put("/clinics/{id}", h.update)
	r.Delete("/clinics/{id}", h.delete)
}

type handler struct {
	svc *Service
}

// create godoc
// @Summary      Create a clinic
// @Tags         clinics
// @Accept       json
// @Produce      json
// @Param        clinic  body      CreateInput  true  "Clinic to create"
// @Success      201     {object}  clinicResponse
// @Failure      400     {object}  errorResponse
// @Failure      409     {object}  errorResponse
// @Router       /clinics [post]
func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}

	c, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toClinicResponse(c))
}

// get godoc
// @Summary      Get a clinic by id
// @Tags         clinics
// @Produce      json
// @Param        id   path      string  true  "Clinic ID"
// @Success      200  {object}  clinicResponse
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
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

	writeJSON(w, http.StatusOK, toClinicResponse(c))
}

// update godoc
// @Summary      Update a clinic (partial)
// @Tags         clinics
// @Accept       json
// @Produce      json
// @Param        id      path      string       true  "Clinic ID"
// @Param        clinic  body      UpdateInput  true  "Fields to update"
// @Success      200     {object}  clinicResponse
// @Failure      400     {object}  errorResponse
// @Failure      404     {object}  errorResponse
// @Router       /clinics/{id} [put]
func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var in UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}

	c, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toClinicResponse(c))
}

// delete godoc
// @Summary      Soft-delete a clinic
// @Tags         clinics
// @Param        id   path  string  true  "Clinic ID"
// @Success      204
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
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

func parseID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return "", false
	}
	return id, true
}

func writeDomainError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "request validation failed", validationErr.Fields)
	case errors.Is(err, ErrDocumentImmutable):
		writeError(w, http.StatusBadRequest, "DOCUMENT_IMMUTABLE", "document cannot be changed after creation", nil)
	case errors.Is(err, ErrDocumentExists):
		writeError(w, http.StatusConflict, "DOCUMENT_ALREADY_EXISTS", "document already belongs to another clinic", nil)
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "CLINIC_NOT_FOUND", "clinic not found", nil)
	default:
		slog.Error("unclassified clinic repository error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", nil)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string, fields map[string]string) {
	writeJSON(w, status, errorResponse{Error: code, Message: message, Fields: fields})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
