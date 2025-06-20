package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"
	"net"
)

type ProbeStatus struct {
	Ready   int32 // 0 = not ready, 1 = ready
	Healthy int32 // 0 = not healthy, 1 = healthy
}

// Dodaj nowe pole "IP" do struktury ServerInfo
type ServerInfo struct {
	Hostname  string   `json:"hostname"`
	OS        string   `json:"os"`
	GoVersion string   `json:"go_version"`
	CPUs      int      `json:"cpus"`
	Env       []string `json:"env"`
	IP        string   `json:"ip"`
}

var probeStatus = ProbeStatus{
	Ready:   1,
	Healthy: 1,
}

// Funkcja pomocnicza do uzyskania adresu IP
func getServerIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("Error fetching IP addresses: %v", err)
		return "unknown"
	}

	for _, addr := range addrs {
		// Sprawdź, czy jest to adres IP (IPv4) i czy nie jest to adres lokalny
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return "unknown"
}

func main() {
	if _, err := os.Stat("templates"); os.IsNotExist(err) {
		err := os.MkdirAll("templates", 0755)
		if err != nil {
			log.Fatalf("Cannot create templates directory: %v", err)
		}
	}

	if _, err := os.Stat("static"); os.IsNotExist(err) {
		err := os.MkdirAll("static", 0755)
		if err != nil {
			log.Fatalf("Cannot create static directory: %v", err)
		}
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		// Pobierz hostname
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}

		// Uzupełnianie danych ServerInfo, łącznie z IP
		serverInfo := ServerInfo{
			Hostname:  hostname,
			OS:        runtime.GOOS,
			GoVersion: runtime.Version(),
			CPUs:      runtime.NumCPU(),
			Env:       os.Environ(),
			IP:        getServerIP(),
		}

		// Przygotowanie danych dla szablonu
		data := struct {
			Endpoints  map[string]string
			Year       int
			ServerInfo ServerInfo
		}{
			Endpoints: map[string]string{
				"/health":         "Kubernetes liveness probe",
				"/ready":          "Kubernetes readiness probe",
				"/toggle/ready":   "Toggle readiness status",
				"/toggle/healthy": "Toggle health status",
			},
			Year:       time.Now().Year(),
			ServerInfo: serverInfo,
		}

		// Wczytaj szablon
		tmplPath := filepath.Join("templates", "index.html")
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			log.Printf("Error loading template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Renderuj dane w szablonie
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	// Kubernetes liveness probe
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		isHealthy := atomic.LoadInt32(&probeStatus.Healthy) == 1
		response := map[string]interface{}{
			"status": isHealthy,
			"message": "Healthy",
		}

		if !isHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
			response["message"] = "Unhealthy"
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(response)
	})

	// Kubernetes readiness probe
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		isReady := atomic.LoadInt32(&probeStatus.Ready) == 1
		response := map[string]interface{}{
			"status": isReady,
			"message": "Ready",
		}

		if !isReady {
			w.WriteHeader(http.StatusServiceUnavailable)
			response["message"] = "Not Ready"
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(response)
	})

	// Toggle readiness status
	http.HandleFunc("/toggle/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		current := atomic.LoadInt32(&probeStatus.Ready)
		var newStatus int32
		var message string

		if current == 1 {
			newStatus = 0
			message = "Readiness probe disabled. Pod will be removed from service."
			atomic.StoreInt32(&probeStatus.Ready, newStatus)
		} else {
			newStatus = 1
			message = "Readiness probe enabled. Pod will receive traffic."
			atomic.StoreInt32(&probeStatus.Ready, newStatus)
		}

		response := map[string]interface{}{
			"previous_status": current == 1,
			"current_status": newStatus == 1,
			"message": message,
		}

		json.NewEncoder(w).Encode(response)
	})

	// Toggle health status
	http.HandleFunc("/toggle/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		current := atomic.LoadInt32(&probeStatus.Healthy)
		var newStatus int32
		var message string

		if current == 1 {
			newStatus = 0
			message = "Liveness probe disabled. Pod will be restarted by Kubernetes."
			atomic.StoreInt32(&probeStatus.Healthy, newStatus)
		} else {
			newStatus = 1
			message = "Liveness probe enabled. Pod is considered healthy."
			atomic.StoreInt32(&probeStatus.Healthy, newStatus)
		}

		response := map[string]interface{}{
			"previous_status": current == 1,
			"current_status": newStatus == 1,
			"message": message,
		}

		json.NewEncoder(w).Encode(response)
	})

	// Start the server
	port := "8080"
	log.Printf("Starting server on :%s", port)
	log.Printf("Kubernetes Demo Application (kuard) is running")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}