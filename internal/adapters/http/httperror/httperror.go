package httperror

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Write(w http.ResponseWriter, status int, code, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Response{Code: code, Message: message})
}

func WriteInternal(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	Write(w, http.StatusInternalServerError, "internal_error", "ocurrio un error, intenta de nuevo")
}
