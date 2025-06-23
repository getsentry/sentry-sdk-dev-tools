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

// PageData holds the data for template rendering
type PageData struct {
	FormattedEnvelope string
	Error             string
	HasResult         bool
	Envelope          string
}

// formatEnvelope takes a Sentry envelope and beautifies each JSON line
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

		result.WriteString(fmt.Sprintf("=== Line %d ===\n", lineNum))
		result.WriteString(string(formatted))
		result.WriteString("\n\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading envelope: %v", err)
	}

	return result.String(), nil
}

// homeHandler serves the main form page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{}

	if r.Method == "POST" {
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
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sentry Envelope Formatter</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            font-weight: 600;
            color: #555;
        }
        textarea {
            width: 100%;
            min-height: 200px;
            padding: 12px;
            border: 2px solid #ddd;
            border-radius: 4px;
            font-family: Monaco, 'Courier New', monospace;
            font-size: 13px;
            resize: vertical;
            box-sizing: border-box;
        }
        textarea:focus {
            outline: none;
            border-color: #007acc;
        }
        .submit-container {
            text-align: center;
            margin-top: 20px;
        }
        button {
            background: linear-gradient(135deg, #007acc, #005a9e);
            color: white;
            padding: 15px 32px;
            border: none;
            border-radius: 6px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            box-shadow: 0 2px 8px rgba(0, 122, 204, 0.3);
            min-width: 160px;
        }
        button:hover {
            background: linear-gradient(135deg, #005a9e, #004080);
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(0, 122, 204, 0.4);
        }
        button:active {
            transform: translateY(0);
            box-shadow: 0 2px 6px rgba(0, 122, 204, 0.3);
        }
        .error {
            background: #fee;
            color: #c33;
            padding: 12px;
            border-radius: 4px;
            margin: 20px 0;
            border: 1px solid #fcc;
        }
        .result {
            margin-top: 30px;
        }
        .result h2 {
            color: #333;
            border-bottom: 2px solid #007acc;
            padding-bottom: 10px;
        }
        .formatted-output {
            background: #f8f8f8;
            border: 1px solid #ddd;
            border-radius: 4px;
            padding: 20px;
            font-family: Monaco, 'Courier New', monospace;
            font-size: 13px;
            white-space: pre-wrap;
            overflow-x: auto;
            max-height: 600px;
            overflow-y: auto;
        }
        .instructions {
            background: #e8f4f8;
            padding: 15px;
            border-radius: 4px;
            margin-bottom: 20px;
            border-left: 4px solid #007acc;
        }
        .instructions h3 {
            margin-top: 0;
            color: #005a9e;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔧 Sentry Envelope Formatter</h1>
        
        <div class="instructions">
            <h3>How to use:</h3>
            <p>Paste your Sentry envelope data into the text area below. A Sentry envelope contains multiple JSON objects on separate lines. This tool will parse and beautify each JSON object for easier reading.</p>
        </div>

        <form method="POST">
            <div class="form-group">
                <label for="envelope">Paste your Sentry envelope:</label>
                <textarea name="envelope" id="envelope" placeholder="Paste your Sentry envelope here...
Example:
{&quot;event_id&quot;:&quot;12345&quot;,&quot;sent_at&quot;:&quot;2023-01-01T00:00:00Z&quot;}
{&quot;type&quot;:&quot;event&quot;,&quot;content_type&quot;:&quot;application/json&quot;}
{&quot;message&quot;:&quot;Hello World&quot;}">{{.Envelope}}</textarea>
            </div>
            <div class="submit-container">
                <button type="submit">🚀 Format Envelope</button>
            </div>
        </form>

        {{if .Error}}
        <div class="error">
            <strong>Error:</strong> {{.Error}}
        </div>
        {{end}}

        {{if .HasResult}}
        <div class="result">
            <h2>📋 Formatted Envelope</h2>
            <div class="formatted-output">{{.FormattedEnvelope}}</div>
        </div>
        {{end}}
    </div>
</body>
</html>`

	t, err := template.New("index").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func main() {
	http.HandleFunc("/", homeHandler)

	fmt.Println("🚀 Sentry Envelope Formatter running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
