package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"gox/internal/config"
	"gox/internal/proxy"
	"gox/internal/watcher"
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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
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
			r.runAppWithLiveReload(ctx, name, app, col)
		}(appName, app, color)
	}

	wg.Wait()
	fmt.Println("All services stopped.")
	return nil
}

func (r *Runner) runAppWithLiveReload(ctx context.Context, name string, app config.App, color string) {
	appDir := filepath.Join(r.ProjectRoot, app.Path)

	// Start File Watcher
	w, err := watcher.NewWatcher(appDir)
	if err != nil {
		fmt.Printf("[%s] Error starting watcher: %v\n", name, err)
		return
	}
	defer w.Close()

	// Start Proxy Server
	proxyPort := app.Port
	if proxyPort == "" {
		proxyPort = "8080"
	}

	dp := proxy.NewDynamicProxy()
	proxyServer := &http.Server{
		Addr:    ":" + proxyPort,
		Handler: dp,
	}

	go func() {
		fmt.Printf("%s[%s]%s Reverse Proxy listening on :%s\n", color, name, resetColor, proxyPort)
		if err := proxyServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[%s] Proxy error: %v\n", name, err)
		}
	}()
	defer proxyServer.Close()

	var currentCmd *exec.Cmd
	backendPort := 35000 // Starting dynamic port

	buildAndSwap := func() {
		backendPort++
		fmt.Printf("%s[%s]%s Building new version...\n", color, name, resetColor)

		binPath := filepath.Join(r.ProjectRoot, "bin", fmt.Sprintf("%s_tmp.exe", name))
		mainPath := app.Main
		if mainPath == "" {
			mainPath = "."
		}

		buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, mainPath)
		buildCmd.Dir = appDir
		if out, err := buildCmd.CombinedOutput(); err != nil {
			fmt.Printf("%s[%s]%s Build failed: %v\n%s\n", color, name, resetColor, err, string(out))
			return
		}

		runCmd := exec.CommandContext(ctx, binPath)
		runCmd.Dir = appDir
		runCmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", backendPort))

		stdout, _ := runCmd.StdoutPipe()
		stderr, _ := runCmd.StderrPipe()

		if err := runCmd.Start(); err != nil {
			fmt.Printf("%s[%s]%s Failed to start: %v\n", color, name, resetColor, err)
			return
		}

		var streamWg sync.WaitGroup
		streamWg.Add(2)
		prefix := fmt.Sprintf("%s[%s:%d]%s ", color, name, backendPort, resetColor)
		go r.streamLogs(stdout, prefix, &streamWg)
		go r.streamLogs(stderr, prefix, &streamWg)

		// Give the new process a moment to bind to the port
		time.Sleep(500 * time.Millisecond)

		err = dp.SetTarget(fmt.Sprintf("http://127.0.0.1:%d", backendPort))
		if err != nil {
			fmt.Printf("%s[%s]%s Failed to set proxy target: %v\n", color, name, resetColor, err)
			return
		}
		fmt.Printf("%s[%s]%s Live Reload Complete! Traffic switched to port %d\n", color, name, resetColor, backendPort)

		if currentCmd != nil && currentCmd.Process != nil {
			fmt.Printf("%s[%s]%s Initiating graceful shutdown of old process (PID: %d)...\n", color, name, resetColor, currentCmd.Process.Pid)
			// Terminate gracefully using interrupt
			currentCmd.Process.Signal(os.Interrupt)
		}

		currentCmd = runCmd
	}

	buildAndSwap()

	for {
		select {
		case <-ctx.Done():
			if currentCmd != nil && currentCmd.Process != nil {
				currentCmd.Process.Kill()
			}
			return
		case <-w.OnChange:
			buildAndSwap()
		}
	}
}

func (r *Runner) streamLogs(reader io.Reader, prefix string, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Println(prefix + scanner.Text())
	}
}
