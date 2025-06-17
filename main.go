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
)

type ProbeStatus struct {
	Ready   int32 // 0 = not ready, 1 = ready
	Healthy int32 // 0 = not healthy, 1 = healthy
}

type ServerInfo struct {
	Hostname  string   `json:"hostname"`
	OS        string   `json:"os"`
	GoVersion string   `json:"go_version"`
	CPUs      int      `json:"cpus"`
	Env       []string `json:"env"`
}

var probeStatus = ProbeStatus{
	Ready:   1,
	Healthy: 1,
}

func main() {
	if _, err := os.Stat("templates"); os.IsNotExist(err) {
		err := os.MkdirAll("templates", 0755)
		if err != nil {
			log.Fatalf("Nie można utworzyć katalogu templates: %v", err)
		}
	}

	if _, err := os.Stat("static"); os.IsNotExist(err) {
		err := os.MkdirAll("static", 0755)
		if err != nil {
			log.Fatalf("Nie można utworzyć katalogu static: %v", err)
		}
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		// Przygotowanie danych dla szablonu
		data := struct {
			Endpoints map[string]string
			Year      int
		}{
			Endpoints: map[string]string{
				"/request":        "Wyświetl informacje o żądaniu",
				"/server":         "Wyświetl informacje o serwerze",
				"/health":         "Kubernetes liveness probe",
				"/ready":          "Kubernetes readiness probe",
				"/toggle/ready":   "Przełącz stan gotowości",
				"/toggle/healthy": "Przełącz stan zdrowia",
			},
			Year: time.Now().Year(),
		}

		// Wczytanie szablonu
		tmplPath := filepath.Join("templates", "index.html")
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			log.Printf("Błąd wczytywania szablonu: %v", err)
			http.Error(w, "Błąd wewnętrzny serwera", http.StatusInternalServerError)
			return
		}

		// Ustawienie nagłówka i renderowanie szablonu
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Błąd renderowania szablonu: %v", err)
			http.Error(w, "Błąd wewnętrzny serwera", http.StatusInternalServerError)
		}
	})

	// Request info handler
	http.HandleFunc("/request", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		headers := make(map[string][]string)
		for name, values := range r.Header {
			headers[name] = values
		}

		requestInfo := map[string]interface{}{
			"method": r.Method,
			"url": r.URL.String(),
			"protocol": r.Proto,
			"host": r.Host,
			"remote_address": r.RemoteAddr,
			"headers": headers,
		}

		json.NewEncoder(w).Encode(requestInfo)
	})

	// Server info handler
	http.HandleFunc("/server", func(w http.ResponseWriter, r *http.Request) {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}

		info := ServerInfo{
			Hostname:  hostname,
			OS:        runtime.GOOS,
			GoVersion: runtime.Version(),
			CPUs:      runtime.NumCPU(),
			Env:       os.Environ(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
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
