package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

func TenantResolver(weddingRepo repository.WeddingRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			weddingID := chi.URLParam(r, "weddingId")
			if weddingID == "" {
				notFound(w, "Identificação do casamento não fornecida")
				return
			}

			wedding, err := weddingRepo.FindByID(r.Context(), weddingID)
			if err != nil {
				if err == entity.ErrNotFound {
					notFound(w, "Casamento não encontrado")
					return
				}
				serverError(w)
				return
			}

			if !wedding.Active {
				notFound(w, "Casamento não encontrado")
				return
			}

			ctx := context.WithValue(r.Context(), WeddingIDKey, wedding.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func notFound(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func serverError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"error": "Erro interno do servidor"})
}
