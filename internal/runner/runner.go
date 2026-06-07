package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"gox/internal/config"
)

var colors = []string{
	"\033[36m", // Cyan
	"\033[32m", // Green
	"\033[33m", // Yellow
	"\033[35m", // Magenta
	"\033[34m", // Blue
}
var resetColor = "\033[0m"

type Runner struct {
	Config      *config.GoxConfig
	ProjectRoot string
}

func NewRunner(cfg *config.GoxConfig, root string) *Runner {
	return &Runner{
		Config:      cfg,
		ProjectRoot: root,
	}
}

func (r *Runner) RunDev() error {
	tasks, ok := r.Config.Tasks["dev"]
	if !ok || len(tasks) == 0 {
		return fmt.Errorf("no dev tasks found in gox.yaml")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nReceived interrupt, shutting down gracefully...")
		cancel()
	}()

	var wg sync.WaitGroup
	colorIdx := 0

	for _, task := range tasks {
		appName := task.Run
		app, exists := r.Config.Apps[appName]
		if !exists {
			fmt.Printf("Warning: task refers to unknown app '%s'\n", appName)
			continue
		}

		color := colors[colorIdx%len(colors)]
		colorIdx++

		wg.Add(1)
		go func(name string, app config.App, col string) {
			defer wg.Done()
			r.runProcess(ctx, name, app, col)
		}(appName, app, color)
	}

	wg.Wait()
	fmt.Println("All services stopped.")
	return nil
}

func (r *Runner) runProcess(ctx context.Context, name string, app config.App, color string) {
	appDir := filepath.Join(r.ProjectRoot, app.Path)
	mainPath := app.Main
	if mainPath == "" {
		mainPath = "."
	}

	cmd := exec.CommandContext(ctx, "go", "run", mainPath)
	cmd.Dir = appDir

	// Setup environment variables if needed
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%s", app.Port))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("[%s] Error setting up stdout: %v\n", name, err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("[%s] Error setting up stderr: %v\n", name, err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("[%s] Failed to start: %v\n", name, err)
		return
	}

	prefix := fmt.Sprintf("%s[%s]%s ", color, name, resetColor)
	fmt.Printf("%sStarted (PID: %d)\n", prefix, cmd.Process.Pid)

	var streamWg sync.WaitGroup
	streamWg.Add(2)

	go r.streamLogs(stdout, prefix, &streamWg)
	go r.streamLogs(stderr, prefix, &streamWg)

	streamWg.Wait()

	err = cmd.Wait()
	if err != nil {
		if ctx.Err() != nil {
			fmt.Printf("%sStopped\n", prefix)
		} else {
			fmt.Printf("%sExited with error: %v\n", prefix, err)
		}
	} else {
		fmt.Printf("%sExited cleanly\n", prefix)
	}
}

func (r *Runner) streamLogs(reader io.Reader, prefix string, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Println(prefix + scanner.Text())
	}
}
