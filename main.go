package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

type PageData struct {
	FormattedEnvelope string
	Error             string
	HasResult         bool
	Envelope          string
}

type ResultData struct {
	FormattedEnvelope string
	Error             string
	HasResult         bool
	Envelope          string
	Timestamp         time.Time
}

var templates *template.Template
var resultStore = make(map[string]ResultData)
var resultMutex sync.RWMutex

func formatEnvelope(envelope string) (string, error) {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(envelope))
	lineNum := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		// Skip empty lines
		if line == "" {
			result.WriteString("\n")
			continue
		}

		// Try to parse and format as JSON
		var jsonObj interface{}
		if err := json.Unmarshal([]byte(line), &jsonObj); err != nil {
			return "", fmt.Errorf("line %d is not valid JSON: %v", lineNum, err)
		}

		// Pretty print the JSON
		formatted, err := json.MarshalIndent(jsonObj, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to format JSON on line %d: %v", lineNum, err)
		}
		result.WriteString(string(formatted))
		result.WriteString("\n\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading envelope: %v", err)
	}

	return result.String(), nil
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func storeResult(data ResultData) string {
	id := generateID()
	data.Timestamp = time.Now()

	resultMutex.Lock()
	resultStore[id] = data
	resultMutex.Unlock()

	go cleanupOldResults()

	return id
}

func getResult(id string) (ResultData, bool) {
	resultMutex.RLock()
	data, exists := resultStore[id]
	resultMutex.RUnlock()
	return data, exists
}

func cleanupOldResults() {
	resultMutex.Lock()
	defer resultMutex.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)
	for id, data := range resultStore {
		if data.Timestamp.Before(cutoff) {
			delete(resultStore, id)
		}
	}
}

func initTemplates() {
	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Error parsing templates: ", err)
	}
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{}

	// Check if there's a result ID in the URL parameters
	resultID := r.URL.Query().Get("result")
	if resultID != "" {
		if resultData, exists := getResult(resultID); exists {
			data.FormattedEnvelope = resultData.FormattedEnvelope
			data.Error = resultData.Error
			data.HasResult = resultData.HasResult
			data.Envelope = resultData.Envelope
		}
	}

	w.Header().Set("Content-Type", "text/html")
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	resultData := ResultData{}

	envelope := r.FormValue("envelope")
	resultData.Envelope = envelope // Preserve user input
	if envelope == "" {
		resultData.Error = "Please paste a Sentry envelope"
	} else {
		formatted, err := formatEnvelope(envelope)
		if err != nil {
			resultData.Error = err.Error()
		} else {
			resultData.FormattedEnvelope = formatted
			resultData.HasResult = true
		}
	}

	// Store the result and redirect to prevent resubmission
	resultID := storeResult(resultData)
	http.Redirect(w, r, "/?result="+resultID, http.StatusSeeOther)
}

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Debug:            true,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)

	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	initTemplates()

	mux := http.NewServeMux()
	mux.Handle("GET /", sentryHandler.Handle(http.HandlerFunc(GetHandler)))
	mux.Handle("POST /", sentryHandler.Handle(http.HandlerFunc(PostHandler)))

	log.Print("Running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
