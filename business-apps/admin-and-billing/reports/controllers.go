package reports

import (
	"encoding/json"
	"errors"
	"net/http"
)

func GenerateBill(w http.ResponseWriter, r *http.Request) {
	var req GenerateBillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bills, err := GenerateBillReportService(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrUserDoesNotExist) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(bills)
}
