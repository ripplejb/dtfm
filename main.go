package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/ollama/ollama/api"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Style string `yaml:"style"`
}

const (
	modelName = "deepseek-coder:latest"
	prompt    = `You are a command line expert. Given a task description, provide the exact command or series of commands to accomplish it. Only output the commands, no explanations.

Task: %s

Commands:`
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: dtfm 'task description'")
		fmt.Println("Example: dtfm 'list files in the current folder'")
		os.Exit(1)
	}

	// Get the task description from command line arguments
	task := strings.Join(os.Args[1:], " ")
	// Create Ollama client
	client := api.NewClient(&url.URL{
		Scheme: "http",
		Host:   "localhost:11434",
	}, http.DefaultClient)

	// Prepare the context
	ctx := context.Background()
	// Generate the response
	var response string
	if err := client.Generate(ctx, &api.GenerateRequest{
		Model:  modelName,
		Prompt: fmt.Sprintf(prompt, task),
	}, func(r api.GenerateResponse) error {
		response += r.Response
		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating response: %v\n", err)
		os.Exit(1)
	}

	// Get the executable's directory
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exePath)

	// Load style from config
	configPath := filepath.Join(exeDir, "dtfm_config.yaml")
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file at %s: %v\n", configPath, err)
		os.Exit(1)
	}

	var config Config
	if err := yaml.Unmarshal(configFile, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config file: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	r, _ := glamour.NewTermRenderer(
		glamour.WithStylePath(config.Style),
		glamour.WithWordWrap(80),
	)
	out, err := r.Render(response)
	if err != nil {
		fmt.Println(response) // Fallback to unformatted output if rendering fails
	} else {
		fmt.Print(out)
	}
}
