package dentist

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func RegisterRoutes(r chi.Router, svc *Service) {
	h := &handler{svc: svc}
	r.Post("/clinics/{clinic_id}/dentists", h.create)
	r.Get("/clinics/{clinic_id}/dentists", h.list)
	r.Get("/clinics/{clinic_id}/dentists/{id}", h.get)
	r.Put("/clinics/{clinic_id}/dentists/{id}", h.update)
	r.Delete("/clinics/{clinic_id}/dentists/{id}", h.delete)
}

type handler struct {
	svc *Service
}

// create godoc
// @Summary      Create a dentist in a clinic
// @Tags         dentists
// @Accept       json
// @Produce      json
// @Param        clinic_id  path      string       true  "Clinic ID"
// @Param        dentist    body      CreateInput  true  "Dentist to create"
// @Success      201        {object}  dentistResponse
// @Failure      400        {object}  errorResponse
// @Failure      404        {object}  errorResponse
// @Failure      409        {object}  errorResponse
// @Router       /clinics/{clinic_id}/dentists [post]
func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	clinicID, ok := parseUUID(w, chi.URLParam(r, "clinic_id"))
	if !ok {
		return
	}

	var in CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}

	d, err := h.svc.Create(r.Context(), clinicID, in)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toDentistResponse(d))
}

// get godoc
// @Summary      Get a dentist by id
// @Tags         dentists
// @Produce      json
// @Param        clinic_id  path      string  true  "Clinic ID"
// @Param        id         path      string  true  "Dentist ID"
// @Success      200        {object}  dentistResponse
// @Failure      400        {object}  errorResponse
// @Failure      404        {object}  errorResponse
// @Router       /clinics/{clinic_id}/dentists/{id} [get]
func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	clinicID, id, ok := parsePathIDs(w, r)
	if !ok {
		return
	}

	d, err := h.svc.Get(r.Context(), clinicID, id)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toDentistResponse(d))
}

// update godoc
// @Summary      Update a dentist (partial)
// @Tags         dentists
// @Accept       json
// @Produce      json
// @Param        clinic_id  path      string       true  "Clinic ID"
// @Param        id         path      string       true  "Dentist ID"
// @Param        dentist    body      UpdateInput  true  "Fields to update"
// @Success      200        {object}  dentistResponse
// @Failure      400        {object}  errorResponse
// @Failure      404        {object}  errorResponse
// @Failure      409        {object}  errorResponse
// @Router       /clinics/{clinic_id}/dentists/{id} [put]
func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	clinicID, id, ok := parsePathIDs(w, r)
	if !ok {
		return
	}

	var in UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}

	d, err := h.svc.Update(r.Context(), clinicID, id, in)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toDentistResponse(d))
}

// delete godoc
// @Summary      Soft-delete a dentist
// @Tags         dentists
// @Param        clinic_id  path  string  true  "Clinic ID"
// @Param        id         path  string  true  "Dentist ID"
// @Success      204
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /clinics/{clinic_id}/dentists/{id} [delete]
func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	clinicID, id, ok := parsePathIDs(w, r)
	if !ok {
		return
	}

	if err := h.svc.Delete(r.Context(), clinicID, id); err != nil {
		writeDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// list godoc
// @Summary      List dentists of a clinic (paginated)
// @Tags         dentists
// @Produce      json
// @Param        clinic_id  path      string  true   "Clinic ID"
// @Param        limit      query     int     false  "Page size (default 20, max 100)"
// @Param        offset     query     int     false  "Offset (default 0)"
// @Success      200        {object}  listResponse
// @Failure      400        {object}  errorResponse
// @Failure      404        {object}  errorResponse
// @Router       /clinics/{clinic_id}/dentists [get]
func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	clinicID, ok := parseUUID(w, chi.URLParam(r, "clinic_id"))
	if !ok {
		return
	}

	params := ListParams{
		Limit:  clampLimit(parseQueryInt(r, "limit")),
		Offset: clampOffset(parseQueryInt(r, "offset")),
	}

	result, err := h.svc.List(r.Context(), clinicID, params)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toListResponse(result, params))
}

func parseQueryInt(r *http.Request, key string) int {
	v, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return 0
	}
	return v
}

func parsePathIDs(w http.ResponseWriter, r *http.Request) (clinicID, id string, ok bool) {
	clinicID, ok = parseUUID(w, chi.URLParam(r, "clinic_id"))
	if !ok {
		return "", "", false
	}
	id, ok = parseUUID(w, chi.URLParam(r, "id"))
	if !ok {
		return "", "", false
	}
	return clinicID, id, true
}

func parseUUID(w http.ResponseWriter, raw string) (string, bool) {
	if _, err := uuid.Parse(raw); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return "", false
	}
	return raw, true
}

func writeDomainError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "request validation failed", validationErr.Fields)
	case errors.Is(err, ErrClinicNotFound):
		writeError(w, http.StatusNotFound, "CLINIC_NOT_FOUND", "clinic not found", nil)
	case errors.Is(err, ErrEmailExists):
		writeError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "email already belongs to another dentist in this clinic", nil)
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "DENTIST_NOT_FOUND", "dentist not found", nil)
	default:
		slog.Error("unclassified dentist repository error", "error", err)
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
