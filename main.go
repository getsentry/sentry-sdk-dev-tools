package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type PageData struct {
	FormattedEnvelope string
	Error             string
	HasResult         bool
	Envelope          string
}

var templates *template.Template

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

func initTemplates() {
	var err error
	templates, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Error parsing templates: ", err)
	}
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{}

	w.Header().Set("Content-Type", "text/html")
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{}

	envelope := r.FormValue("envelope")
	data.Envelope = envelope // Preserve user input
	if envelope == "" {
		data.Error = "Please paste a Sentry envelope"
	} else {
		formatted, err := formatEnvelope(envelope)
		if err != nil {
			data.Error = err.Error()
		} else {
			data.FormattedEnvelope = formatted
			data.HasResult = true
		}
	}

	w.Header().Set("Content-Type", "text/html")
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	initTemplates()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", GetHandler)
	mux.HandleFunc("POST /", PostHandler)

	log.Print("Running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
