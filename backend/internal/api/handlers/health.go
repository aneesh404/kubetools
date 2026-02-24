package handlers

import "net/http"

func Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only GET is supported")
		return
	}
	WriteSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}
