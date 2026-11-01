package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ldmejia/PruebaPivote/internal/adapters/http/httperror"
	"github.com/ldmejia/PruebaPivote/internal/adapters/http/middlewares"
)

type HttpHandler struct {
	secret []byte
	ttl    time.Duration
}

func NewHttpHandler(secret []byte, ttl time.Duration) *HttpHandler {
	return &HttpHandler{secret: secret, ttl: ttl}
}

func (h *HttpHandler) IssueToken(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.UserID) == "" {
		httperror.Write(w, http.StatusBadRequest, "invalid_request", "user_id es requerido")
		return
	}
	token, err := middlewares.IssueToken(h.secret, req.UserID, h.ttl)
	if err != nil {
		httperror.WriteInternal(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(TokenResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: int(h.ttl.Seconds())})
}

func (h *HttpHandler) Me(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(map[string]string{"user_id": middlewares.UserID(r.Context())})
}
