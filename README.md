# Sentry Envelope Formatter

A simple Go web application that allows you to paste Sentry envelope data and displays it in a beautified, readable format.

## Quick Start

1. **Run the application:**
   ```bash
   go run main.go
   ```

2. **Open your browser:**
   Navigate to [http://localhost:8080](http://localhost:8080)

3. **Paste your envelope:**
   Copy your Sentry envelope data and paste it into the text area

4. **Format and view:**
   Click "Format Envelope" to see the beautified output

## Example Usage

Input (raw envelope):
```
{"event_id":"12345","sent_at":"2023-01-01T00:00:00Z","dsn":"https://key@sentry.io/project"}
{"type":"event","content_type":"application/json"}
{"message":"Hello World","level":"info","timestamp":"2023-01-01T00:00:00Z"}
```

Output (formatted):
```
=== Line 1 ===
{
  "event_id": "12345",
  "sent_at": "2023-01-01T00:00:00Z",
  "dsn": "https://key@sentry.io/project"
}

=== Line 2 ===
{
  "type": "event",
  "content_type": "application/json"
}

=== Line 3 ===
{
  "message": "Hello World",
  "level": "info",
  "timestamp": "2023-01-01T00:00:00Z"
}
```

## Development

### Requirements

- Go 1.24+ (as specified in go.mod)

### Project Structure

```
envelope-formatter/
├── main.go          # Main application with HTTP handlers
├── go.mod           # Go module definition
└── README.md        # This file
```

### Building

```bash
# Build the application
go build -o envelope-formatter main.go

# Run the built binary
./envelope-formatter
```
