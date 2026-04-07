package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/thnkbig/falcoclaw-hermes/pkg/falcoclaw"
)

var (
	listenAddr   string
	falcoclawURL string
	apiKey       string
)

func init() {
	flag.StringVar(&listenAddr, "listen", "0.0.0.0:2805", "Address to listen on")
	flag.StringVar(&falcoclawURL, "falcoclaw", "http://localhost:2804", "FalcoClaw server URL")
	flag.StringVar(&apiKey, "apikey", "", "FalcoClaw API key")
}

func main() {
	flag.Parse()
	client := falcoclaw.NewClient(falcoclawURL, apiKey)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/webhooks/falco", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var event falcoclaw.Alert
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			log.Printf("ERROR decode: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("INFO Falco alert: %s [%s] %s", event.Priority, event.FalcoRule, event.Hostname)
		alerts, err := client.QueryAlerts(&falcoclaw.QueryOpts{
			Rule:     event.FalcoRule,
			Hostname: event.Hostname,
			Limit:    1,
		})
		if err != nil {
			log.Printf("WARN QueryAlerts: %v", err)
		} else if len(alerts) > 0 {
			log.Printf("INFO FalcoClaw matched rule=%s action=%s",
				alerts[0].FalcoRule, alerts[0].ActionTaken)
		}
		w.WriteHeader(http.StatusAccepted)
	})

	http.HandleFunc("/webhooks/falcoclaw", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			AlertID  string                 `json:"alert_id"`
			Rule     string                 `json:"rule"`
			Priority string                 `json:"priority"`
			Hostname string                 `json:"hostname"`
			Context  map[string]interface{} `json:"context"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("ERROR decode: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("INFO FalcoClaw investigation: alert=%s rule=%s host=%s",
			payload.AlertID, payload.Rule, payload.Hostname)
		w.WriteHeader(http.StatusAccepted)
	})

	log.Printf("Hermes webhook server listening on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
