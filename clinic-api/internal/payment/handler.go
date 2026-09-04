package payment

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/josuesantos1/desafio/pkg/problem"
)

func RegisterRoutes(r chi.Router, svc *Service) {
	h := &handler{svc: svc}
	r.Post("/payments", h.create)
	r.Get("/payments/{id}", h.get)
}

type handler struct {
	svc *Service
}

// @Summary      Create a Pix payment
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header    string       true  "Idempotency key"
// @Param        payment          body      CreateInput  true  "Payment to create"
// @Success      201              {object}  paymentResponse
// @Success      200              {object}  paymentResponse "replay of an existing payment"
// @Failure      400              {object}  problem.Details
// @Failure      404              {object}  problem.Details
// @Failure      409              {object}  problem.Details
// @Router       /payments [post]
func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	if strings.TrimSpace(key) == "" {
		problem.Write(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key header is required", nil)
		return
	}

	var in CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		problem.Write(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}

	p, created, err := h.svc.Create(r.Context(), key, in)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	problem.WriteJSON(w, status, toPaymentResponse(p))
}

// @Summary      Get a payment by id
// @Tags         payments
// @Produce      json
// @Param        id   path      string  true  "Payment ID"
// @Success      200  {object}  paymentResponse
// @Failure      400  {object}  problem.Details
// @Failure      404  {object}  problem.Details
// @Router       /payments/{id} [get]
func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		problem.Write(w, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID", nil)
		return
	}

	p, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	problem.WriteJSON(w, http.StatusOK, toPaymentResponse(p))
}

func writeDomainError(w http.ResponseWriter, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		problem.Write(w, http.StatusBadRequest, "VALIDATION_ERROR", "request validation failed", validationErr.Fields)
	case errors.Is(err, ErrClinicNotFound):
		problem.Write(w, http.StatusNotFound, "CLINIC_NOT_FOUND", "clinic not found", nil)
	case errors.Is(err, ErrClinicNotActive):
		problem.Write(w, http.StatusConflict, "CLINIC_NOT_ACTIVE", "clinic must be active (have an administrator and a legal representative) to receive payments", nil)
	case errors.Is(err, ErrDentistNotFound):
		problem.Write(w, http.StatusNotFound, "DENTIST_NOT_FOUND", "dentist not found", nil)
	case errors.Is(err, ErrIdempotencyKeyConflict):
		problem.Write(w, http.StatusConflict, "IDEMPOTENCY_KEY_CONFLICT", "idempotency key already used with a different payload", nil)
	case errors.Is(err, ErrNotFound):
		problem.Write(w, http.StatusNotFound, "PAYMENT_NOT_FOUND", "payment not found", nil)
	default:
		slog.Error("unclassified payment repository error", "error", err)
		problem.Write(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", nil)
	}
}
