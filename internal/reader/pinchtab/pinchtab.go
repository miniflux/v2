// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package pinchtab // import "miniflux.app/v2/internal/reader/pinchtab"

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"miniflux.app/v2/internal/config"
)

const (
	httpRequestTimeout  = 60 * time.Second
	healthCheckInterval = 500 * time.Millisecond
	healthCheckMaxWait  = 30 * time.Second
	shutdownTimeout     = 5 * time.Second
)

var activeProcessCount atomic.Int64

func ActiveProcessCount() int64 {
	return activeProcessCount.Load()
}

// RenderPage starts an ephemeral pinchtab subprocess in Dashboard mode,
// navigates to the given URL (Chrome is auto-started on the first navigate),
// extracts the rendered text content, and gracefully shuts down everything.
//
// proxyURL is optional: when non-empty it is forwarded to Chrome via
// --proxy-server so that the page fetch goes through the specified proxy.
//
// feedID isolates browser state (cookies, localStorage) per feed so that
// login sessions are preserved across fetches but different feeds cannot
// leak state to each other.
func RenderPage(pageURL, proxyURL string, feedID int64) (string, error) {
	port, err := findFreePort()
	if err != nil {
		return "", fmt.Errorf("pinchtab: %w", err)
	}

	cmd, err := startSubprocess(port, proxyURL, feedID)
	if err != nil {
		return "", fmt.Errorf("pinchtab: %w", err)
	}
	// Always ensure cleanup: graceful shutdown first, force kill as fallback.
	defer stopSubprocess(cmd, port)

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	if err := waitForReady(baseURL); err != nil {
		return "", fmt.Errorf("pinchtab: %w", err)
	}

	client := &http.Client{Timeout: httpRequestTimeout}

	// Dashboard mode: POST /tab with url triggers automatic Chrome startup
	// and tab creation in one call. No manual instance management needed.
	tabID, err := createTab(client, baseURL, pageURL)
	if err != nil {
		return "", fmt.Errorf("pinchtab: failed to create tab for %q: %w", pageURL, err)
	}

	text, err := getTabText(client, baseURL, tabID)
	if err != nil {
		return "", fmt.Errorf("pinchtab: failed to get text for %q: %w", pageURL, err)
	}

	return text, nil
}

func startSubprocess(port int, proxyURL string, feedID int64) (*exec.Cmd, error) {
	binaryPath := config.Opts.PinchTabBinaryPath()
	cmd := exec.Command(binaryPath)

	// Per-feed state directory provides browser state isolation (cookies,
	// localStorage) between feeds. Each feed gets its own Chrome profile
	// under {stateDir}/profiles/default/.
	stateDir := filepath.Join(os.TempDir(), "miniflux-pinchtab-state", fmt.Sprintf("feed-%d", feedID))

	// Container environments require --no-sandbox and --disable-dev-shm-usage
	// for Chrome to start. These are always included, with --proxy-server
	// appended when a proxy is configured.
	chromeFlags := "--no-sandbox --disable-dev-shm-usage"
	if proxyURL != "" {
		chromeFlags += " --proxy-server=" + proxyURL
	}

	env := append(os.Environ(),
		"PINCHTAB_STEALTH=full",
		"PINCHTAB_PORT="+strconv.Itoa(port),
		"PINCHTAB_STATE_DIR="+stateDir,
		"PINCHTAB_HEADLESS=true",
		"CHROME_FLAGS="+chromeFlags,
	)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	slog.Debug("pinchtab: starting ephemeral subprocess",
		slog.String("binary", binaryPath),
		slog.Int("port", port),
		slog.String("proxy_url", proxyURL),
	)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start %q: %w", binaryPath, err)
	}
	activeProcessCount.Add(1)

	return cmd, nil
}

// stopSubprocess gracefully shuts down pinchtab via POST /shutdown so it can
// clean up Chrome child processes. Falls back to Kill if shutdown fails.
func stopSubprocess(cmd *exec.Cmd, port int) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	defer activeProcessCount.Add(-1)

	shutdownURL := fmt.Sprintf("http://127.0.0.1:%d/shutdown", port)
	client := &http.Client{Timeout: shutdownTimeout}
	resp, err := client.Post(shutdownURL, "application/json", nil)
	if err == nil {
		resp.Body.Close()
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case <-done:
		return
	case <-time.After(shutdownTimeout):
		slog.Warn("pinchtab: shutdown timed out, force killing")
		_ = cmd.Process.Kill()
		// Reap zombie process.
		<-done
	}
}

func waitForReady(baseURL string) error {
	healthURL := baseURL + "/health"
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(healthCheckMaxWait)

	for time.Now().Before(deadline) {
		resp, err := client.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(healthCheckInterval)
	}

	return fmt.Errorf("health check timed out after %s", healthCheckMaxWait)
}

func findFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("unable to find free port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port, nil
}

type tabResponse struct {
	TabID string `json:"tabId"`
}

type textResponse struct {
	Text string `json:"text"`
}

// createTab uses the Dashboard-mode shortcut: POST /tab with {"action":"new","url":"..."}
// triggers automatic Chrome startup and tab creation.
func createTab(client *http.Client, baseURL, pageURL string) (string, error) {
	payload := fmt.Sprintf(`{"action":"new","url":%s}`, quote(pageURL))
	resp, err := client.Post(baseURL+"/tab", "application/json", strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var result tabResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode tab response: %w", err)
	}

	return result.TabID, nil
}

func getTabText(client *http.Client, baseURL, tabID string) (string, error) {
	url := fmt.Sprintf("%s/text?tabId=%s", baseURL, tabID)
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var result textResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode text response: %w", err)
	}

	return result.Text, nil
}

func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// ChromiumProcessCount scans /proc to count running chromium-browser processes.
// Returns 0 on non-Linux systems or if /proc is unavailable.
func ChromiumProcessCount() int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid := entry.Name()
		if len(pid) == 0 || pid[0] < '1' || pid[0] > '9' {
			continue
		}
		cmdline, err := os.ReadFile("/proc/" + pid + "/cmdline")
		if err != nil {
			continue
		}
		// /proc/PID/cmdline uses null bytes as separators; the executable is the first field.
		if nullIdx := strings.IndexByte(string(cmdline), 0); nullIdx > 0 {
			exe := string(cmdline[:nullIdx])
			if strings.Contains(exe, "chromium") || strings.Contains(exe, "chrome") {
				count++
			}
		}
	}
	return count
}
