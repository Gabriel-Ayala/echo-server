package echo

import (
	"encoding/json"
	"io"
	"net/http"
)

type EchoResponse struct {
	Headers map[string][]string `json:"headers"`
	Params  map[string][]string `json:"params"`
	Body    string              `json:"body"`
	Path    string              `json:"path"`
}

func NewHandler() http.Handler {
	return http.HandlerFunc(echoHandler)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	headers := r.Header.Clone()
	if r.Host != "" {
		headers.Set("Host", r.Host)
	}

	resp := EchoResponse{
		Headers: headers,
		Params:  r.URL.Query(),
		Body:    string(bodyBytes),
		Path:    r.URL.Path,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
