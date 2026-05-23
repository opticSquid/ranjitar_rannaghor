package stats

import (
	"encoding/json"
	"net/http"
)

func GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := GetDashboardStatsService(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(stats)
}

func GetAnalyticsStats(w http.ResponseWriter, r *http.Request) {
	stats, err := GetAnalyticsStatsService(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(stats)
}
