package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rsdenck/nux/internal/output"
	"github.com/spf13/cobra"
)

var znetCmd = &cobra.Command{
	Use:   "znet",
	Short: "ZeroNet P2P network management",
	Long:  `Manage ZeroNet decentralized P2P networks. Browse sites, manage peers, and monitor the ZeroNet daemon.`,
}

var znetStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start ZeroNet daemon",
	Long:  `Start the ZeroNet background process on port 43110. Auto-installs if not present.`,
	Run: func(cmd *cobra.Command, args []string) {
		if isZeroNetRunning() {
			output.NewInfo("ZeroNet is already running").Print()
			return
		}
		home, _ := os.UserHomeDir()
		znetDir := filepath.Join(home, ".nux", "znet")
		znetPy := filepath.Join(znetDir, "zeronet.py")

		if _, err := os.Stat(znetPy); os.IsNotExist(err) {
			output.NewInfo("ZeroNet not found. Installing...").Print()
			installZeroNet(znetDir)
			if _, err := os.Stat(znetPy); os.IsNotExist(err) {
				output.NewError("Auto-install failed. Check 'nux znet doctor' for details.", "ZNET_INSTALL_ERR").Print()
				return
			}
		}

		cmdStart := exec.Command("python3", znetPy)
		cmdStart.Dir = znetDir
		if err := cmdStart.Start(); err != nil {
			output.NewError(fmt.Sprintf("Failed to start ZeroNet: %s", err.Error()), "ZNET_START_ERR").Print()
			return
		}
		pidFile := filepath.Join(znetDir, ".pid")
		os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmdStart.Process.Pid)), 0644)
		output.NewInfo(fmt.Sprintf("ZeroNet started (PID: %d)", cmdStart.Process.Pid)).Print()
		output.NewInfo("Web UI: http://127.0.0.1:43110").Print()
	},
}

var znetStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop ZeroNet daemon",
	Run: func(cmd *cobra.Command, args []string) {
		if !isZeroNetRunning() {
			output.NewInfo("ZeroNet is not running").Print()
			return
		}
		home, _ := os.UserHomeDir()
		pidFile := filepath.Join(home, ".nux", "znet", ".pid")
		pidBytes, err := os.ReadFile(pidFile)
		if err == nil {
			pid := strings.TrimSpace(string(pidBytes))
			exec.Command("kill", pid).Run()
			os.Remove(pidFile)
			output.NewInfo(fmt.Sprintf("ZeroNet (PID %s) stopped", pid)).Print()
			return
		}
		exec.Command("pkill", "-f", "zeronet.py").Run()
		output.NewInfo("ZeroNet stopped").Print()
	},
}

var znetStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show ZeroNet daemon status",
	Run: func(cmd *cobra.Command, args []string) {
		running := isZeroNetRunning()
		home, _ := os.UserHomeDir()
		znetDir := filepath.Join(home, ".nux", "znet")
		installed := false
		if _, err := os.Stat(filepath.Join(znetDir, "zeronet.py")); err == nil {
			installed = true
		}

		headers := []string{"KEY", "VALUE"}
		var rows [][]string
		if installed {
			rows = append(rows, []string{"installed", "yes"})
			rows = append(rows, []string{"path", znetDir})
		} else {
			rows = append(rows, []string{"installed", "no"})
		}
		if running {
			rows = append(rows, []string{"running", "yes"})
			rows = append(rows, []string{"web_ui", "http://127.0.0.1:43110"})
		} else {
			rows = append(rows, []string{"running", "no"})
		}
		output.PrintCompactTable(headers, rows)
	},
}

var znetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ZeroNet sites and connected peers",
	Run: func(cmd *cobra.Command, args []string) {
		if !isZeroNetRunning() {
			output.NewInfo("ZeroNet is not running. Start it with 'nux znet start'").Print()
			return
		}
		sites, err := fetchZeroNetSites()
		if err != nil {
			output.NewError(fmt.Sprintf("Failed to query ZeroNet: %s", err.Error()), "ZNET_QUERY_ERR").Print()
			return
		}
		if len(sites) == 0 {
			output.NewInfo("No sites found in ZeroNet").Print()
			return
		}
		headers := []string{"ADDRESS", "PEERS", "SIZE", "CONTENT UPDATED"}
		var rows [][]string
		for _, s := range sites {
			rows = append(rows, []string{
				s.Address,
				fmt.Sprintf("%d", s.Peers),
				fmt.Sprintf("%.1f MB", float64(s.Size)/1048576),
				s.ContentUpdated,
			})
		}
		output.PrintCompactTable(headers, rows)
	},
}

var znetConnectCmd = &cobra.Command{
	Use:   "connect [site-address]",
	Short: "Connect to a ZeroNet site",
	Long:  `Open a specific ZeroNet site. If no address is given, opens the ZeroNet Web UI.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !isZeroNetRunning() {
			output.NewInfo("Starting ZeroNet first...")
			znetStartCmd.Run(cmd, []string{})
		}
		if len(args) == 0 {
			output.NewInfo("ZeroNet Web UI: http://127.0.0.1:43110").Print()
			exec.Command("xdg-open", "http://127.0.0.1:43110").Start()
			return
		}
		addr := args[0]
		url := fmt.Sprintf("http://127.0.0.1:43110/%s", addr)
		output.NewInfo(fmt.Sprintf("Opening ZeroNet site: %s", url)).Print()
		exec.Command("xdg-open", url).Start()
	},
}

var znetDisconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from ZeroNet (stop daemon)",
	Run: func(cmd *cobra.Command, args []string) {
		znetStopCmd.Run(cmd, args)
	},
}

var znetPeersCmd = &cobra.Command{
	Use:   "peers",
	Short: "Show connected peers count",
	Run: func(cmd *cobra.Command, args []string) {
		if !isZeroNetRunning() {
			output.NewInfo("ZeroNet is not running").Print()
			return
		}
		sites, err := fetchZeroNetSites()
		if err != nil {
			output.NewError(fmt.Sprintf("Failed to query peers: %s", err.Error()), "ZNET_PEERS_ERR").Print()
			return
		}
		totalPeers := 0
		for _, s := range sites {
			totalPeers += s.Peers
		}
		output.NewInfo(fmt.Sprintf("Total peers: %d across %d sites", totalPeers, len(sites))).Print()
	},
}

var znetSitesCmd = &cobra.Command{
	Use:   "sites",
	Short: "List ZeroNet sites and connected peers",
	Run: func(cmd *cobra.Command, args []string) {
		znetListCmd.Run(cmd, args)
	},
}

var znetOpenCmd = &cobra.Command{
	Use:   "open [site-address]",
	Short: "Open a ZeroNet site in browser",
	Long:  `Open a specific ZeroNet site in the browser. If no address is given, opens the ZeroNet Web UI.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		znetConnectCmd.Run(cmd, args)
	},
}

var znetLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show ZeroNet daemon logs",
	Long:  `Show recent log entries from the ZeroNet daemon.`,
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := os.UserHomeDir()
		znetDir := filepath.Join(home, ".nux", "znet")
		logFile := filepath.Join(znetDir, "zeronet.log")
		debugLog := filepath.Join(znetDir, "debug.log")

		var logPath string
		if _, err := os.Stat(logFile); err == nil {
			logPath = logFile
		} else if _, err := os.Stat(debugLog); err == nil {
			logPath = debugLog
		} else {
			output.PrintWarningMessage("No ZeroNet log files found")
			return
		}

		data, err := os.ReadFile(logPath)
		if err != nil {
			output.NewError(fmt.Sprintf("Failed to read log: %s", err.Error()), "ZNET_LOG_ERR").Print()
			return
		}
		lines := strings.Split(string(data), "\n")
		start := 0
		if len(lines) > 50 {
			start = len(lines) - 50
		}
		for _, line := range lines[start:] {
			if strings.TrimSpace(line) != "" {
				fmt.Println(line)
			}
		}
	},
}

var znetDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run ZeroNet health diagnostics",
	Long:  `Comprehensive health check for ZeroNet installation, daemon, port, and connectivity.`,
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := os.UserHomeDir()
		znetDir := filepath.Join(home, ".nux", "znet")
		pidFile := filepath.Join(znetDir, ".pid")

		output.NewInfo("ZeroNet Diagnostics").Print()
		fmt.Println(strings.Repeat("-", 50))

		issues := 0

		znetPy := filepath.Join(znetDir, "zeronet.py")
		if _, err := os.Stat(znetPy); err == nil {
			output.NewInfo("[OK] ZeroNet installed").Print()
		} else {
			output.NewError("[FAIL] ZeroNet not installed", "ZNET_DIAG").Print()
			issues++
		}

		if _, err := os.Stat(pidFile); err == nil {
			pidBytes, _ := os.ReadFile(pidFile)
			pid := strings.TrimSpace(string(pidBytes))
			if _, err := os.FindProcess(atoi(pid)); err == nil {
				output.NewInfo(fmt.Sprintf("[OK] PID file exists: %s", pid)).Print()
			} else {
				output.PrintWarningMessage("[WARN] PID file stale (process not found)")
			}
		} else {
			output.NewInfo("[INFO] No PID file (daemon not running)").Print()
		}

		if isZeroNetRunning() {
			output.NewInfo("[OK] ZeroNet process running").Print()
		} else {
			output.PrintWarningMessage("[WARN] ZeroNet process not running")
		}

		if isPortOpen(43110) {
			output.NewInfo("[OK] Port 43110 is open").Print()
		} else {
			output.PrintWarningMessage("[WARN] Port 43110 is not open")
		}

		if isZeroNetRunning() {
			resp, err := http.Get("http://127.0.0.1:43110")
			if err == nil && resp.StatusCode < 500 {
				output.NewInfo("[OK] Web UI responds (HTTP " + resp.Status + ")").Print()
				resp.Body.Close()
			} else {
				output.PrintWarningMessage("[WARN] Web UI not responding")
			}

			_, err = fetchZeroNetSites()
			if err == nil {
				output.NewInfo("[OK] Stats endpoint accessible").Print()
			} else {
				output.PrintWarningMessage("[WARN] Stats endpoint: " + err.Error())
			}
		}

		if _, err := exec.LookPath("python3"); err == nil {
			output.NewInfo("[OK] python3 found in PATH").Print()
		} else {
			output.NewError("[FAIL] python3 not found", "ZNET_DIAG").Print()
			issues++
		}

		if issues > 0 {
			output.PrintWarningMessage(fmt.Sprintf("Found %d issues requiring attention", issues))
		} else {
			output.NewInfo("All checks passed").Print()
		}
	},
}

type znetSite struct {
	Address        string `json:"address"`
	Peers          int    `json:"peers"`
	Size           int    `json:"size"`
	ContentUpdated string `json:"content_updated"`
}

func installZeroNet(znetDir string) {
	os.MkdirAll(znetDir, 0755)
	url := "https://github.com/ZeroNetX/ZeroNet/archive/refs/heads/master.tar.gz"
	output.NewInfo("Downloading ZeroNet...").Print()
	cmdTar := exec.Command("bash", "-c", fmt.Sprintf(
		"curl -sL '%s' | tar -xzf - -C '%s' --strip-components=1", url, znetDir))
	if out, err := cmdTar.CombinedOutput(); err != nil {
		output.NewError(fmt.Sprintf("Failed to install ZeroNet: %s\n%s", err.Error(), string(out)), "ZNET_INSTALL_ERR").Print()
		return
	}
	output.NewInfo("Applying Python 3 compatibility...").Print()
	exec.Command("bash", "-c",
		fmt.Sprintf("2to3 -w --no-diffs %s/zeronet.py %s/start.py %s/src/main.py %s/src/Config.py 2>/dev/null; "+
			"sed -i 's|python2.7|python3|g' %s/zeronet.py %s/start.py",
			znetDir, znetDir, znetDir, znetDir, znetDir, znetDir)).Run()

	reqFile := filepath.Join(znetDir, "requirements.txt")
	if _, err := os.Stat(reqFile); err == nil {
		output.NewInfo("Installing Python dependencies...").Print()
		cmdPip := exec.Command("pip3", "install", "-r", reqFile)
		if out, err := cmdPip.CombinedOutput(); err != nil {
			output.PrintWarningMessage(fmt.Sprintf("pip install warnings: %s", string(out)))
		}
	}
	output.NewInfo("ZeroNet installed successfully").Print()
}

func isZeroNetRunning() bool {
	out, err := exec.Command("pgrep", "-f", "zeronet.py").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func fetchZeroNetSites() ([]znetSite, error) {
	resp, err := http.Get("http://127.0.0.1:43110/stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var sites []znetSite
	if err := json.Unmarshal(body, &sites); err != nil {
		return nil, err
	}
	return sites, nil
}

func isPortOpen(port int) bool {
	out, err := exec.Command("ss", "-tlnp").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), fmt.Sprintf(":%d", port))
}

func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func init() {
	znetCmd.AddCommand(znetStartCmd)
	znetCmd.AddCommand(znetStopCmd)
	znetCmd.AddCommand(znetStatusCmd)
	znetCmd.AddCommand(znetListCmd)
	znetCmd.AddCommand(znetConnectCmd)
	znetCmd.AddCommand(znetDisconnectCmd)
	znetCmd.AddCommand(znetPeersCmd)
	znetCmd.AddCommand(znetSitesCmd)
	znetCmd.AddCommand(znetOpenCmd)
	znetCmd.AddCommand(znetLogsCmd)
	znetCmd.AddCommand(znetDoctorCmd)
	rootCmd.AddCommand(znetCmd)
}
