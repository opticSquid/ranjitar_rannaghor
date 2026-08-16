package wallet

import (
	"encoding/json"
	"net/http"
)

func RechargeWallet(w http.ResponseWriter, r *http.Request) {
	var req RechargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := rechargeWalletService(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
