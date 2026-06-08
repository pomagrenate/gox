package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

var dockerfileTmpl = `# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /workspace

# Copy workspace configuration
COPY go.work go.work.sum ./
COPY apps/ ./apps/
COPY libs/ ./libs/

# Build the application
WORKDIR /workspace/{{.AppPath}}
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/app {{.MainPath}}

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /bin/app /app

{{if .Port}}EXPOSE {{.Port}}{{end}}
CMD ["/app"]
`

var dockerComposeTmpl = `version: '3.8'

services:
{{- range $name, $app := .Apps }}
  {{ $name }}:
    build:
      context: .
      dockerfile: {{ $app.Path }}/Dockerfile
    container_name: {{ $name }}
    restart: unless-stopped
    {{- if $app.Port }}
    ports:
      - "{{ $app.Port }}:{{ $app.Port }}"
    environment:
      - PORT={{ $app.Port }}
    {{- end }}
{{- end }}
`

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Generate Dockerfiles and docker-compose.yml",
	Long:  `Auto-generates optimized multi-stage Dockerfiles for all apps and a docker-compose.yml to orchestrate them locally.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Generating Docker assets...")

		// Generate Dockerfiles
		dfTemplate, err := template.New("dockerfile").Parse(dockerfileTmpl)
		if err != nil {
			return err
		}

		for name, app := range goxConfig.Apps {
			appDir := filepath.Join(workspaceRoot, filepath.FromSlash(app.Path))
			if _, err := os.Stat(appDir); os.IsNotExist(err) {
				continue // Skip if folder doesn't exist
			}

			dfPath := filepath.Join(appDir, "Dockerfile")
			f, err := os.Create(dfPath)
			if err != nil {
				return err
			}

			mainPath := app.Main
			if mainPath == "" {
				mainPath = "."
			}

			data := struct {
				AppPath  string
				MainPath string
				Port     string
			}{
				AppPath:  filepath.ToSlash(app.Path),
				MainPath: filepath.ToSlash(mainPath),
				Port:     app.Port,
			}

			if err := dfTemplate.Execute(f, data); err != nil {
				f.Close()
				return err
			}
			f.Close()
			fmt.Printf("✅ Created Dockerfile for '%s' at %s\n", name, dfPath)
		}

		// Generate docker-compose.yml
		dcPath := filepath.Join(workspaceRoot, "docker-compose.yml")
		f, err := os.Create(dcPath)
		if err != nil {
			return err
		}
		defer f.Close()

		dcTemplate, err := template.New("compose").Parse(dockerComposeTmpl)
		if err != nil {
			return err
		}

		if err := dcTemplate.Execute(f, goxConfig); err != nil {
			return err
		}
		
		fmt.Printf("✅ Created docker-compose.yml at workspace root\n")
		fmt.Println("\nYou can now run 'docker-compose up --build' to start your entire system!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dockerCmd)
}
