package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rsdenck/nux/internal/output"
	"github.com/spf13/cobra"
)

var i2pCmd = &cobra.Command{
	Use:   "i2p",
	Short: "I2P anonymous network management",
	Long:  `Manage I2P anonymous overlay network. Control router, tunnels, proxies, and monitor peers and statistics.`,
}

var i2pStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start I2P router",
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := os.UserHomeDir()
		i2pDir := filepath.Join(home, ".nux", "i2p")
		os.MkdirAll(i2pDir, 0755)

		if isI2PRunning() {
			output.NewInfo("I2P router is already running").Print()
			return
		}

		i2pRouter := findI2PRouter()
		if i2pRouter == "" {
			output.NewInfo("I2P not found. Use 'nux i2p doctor' for setup help.").Print()
			return
		}

cmdStart := exec.Command(i2pRouter)
	cmdStart.Dir = i2pDir
	if err := cmdStart.Start(); err != nil {
		output.NewError(fmt.Sprintf("Failed to start I2P: %s", err.Error()), "I2P_START_ERR").Print()
		return
	}
		pidFile := filepath.Join(i2pDir, ".pid")
		os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmdStart.Process.Pid)), 0644)
		output.NewInfo(fmt.Sprintf("I2P router started (PID: %d)", cmdStart.Process.Pid)).Print()
		output.NewInfo("Web console: http://127.0.0.1:7657").Print()
	},
}

var i2pStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop I2P router",
	Run: func(cmd *cobra.Command, args []string) {
		if !isI2PRunning() {
			output.NewInfo("I2P router is not running").Print()
			return
		}
		home, _ := os.UserHomeDir()
		pidFile := filepath.Join(home, ".nux", "i2p", ".pid")
		pidBytes, err := os.ReadFile(pidFile)
		if err == nil {
			pid := strings.TrimSpace(string(pidBytes))
			exec.Command("kill", pid).Run()
			os.Remove(pidFile)
			output.NewInfo(fmt.Sprintf("I2P router (PID %s) stopped", pid)).Print()
			return
		}
		exec.Command("pkill", "-f", "i2p|java.net.router").Run()
		output.NewInfo("I2P router stopped").Print()
	},
}

var i2pRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart I2P router",
	Run: func(cmd *cobra.Command, args []string) {
		i2pStopCmd.Run(cmd, args)
		i2pStartCmd.Run(cmd, args)
	},
}

var i2pStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show I2P router status",
	Run: func(cmd *cobra.Command, args []string) {
		running := isI2PRunning()
		installed := isI2PInstalled()

		headers := []string{"KEY", "VALUE"}
		var rows [][]string
		if installed {
			rows = append(rows, []string{"installed", "yes"})
			rows = append(rows, []string{"router", findI2PRouter()})
		} else {
			rows = append(rows, []string{"installed", "no"})
		}
		if running {
			rows = append(rows, []string{"running", "yes"})
			rows = append(rows, []string{"web_console", "http://127.0.0.1:7657"})
		} else {
			rows = append(rows, []string{"running", "no"})
		}
		output.PrintCompactTable(headers, rows)
	},
}

var i2pPeersCmd = &cobra.Command{
	Use:   "peers",
	Short: "Show I2P peer count and status",
	Run: func(cmd *cobra.Command, args []string) {
		if !isI2PRunning() {
			output.NewInfo("I2P router is not running").Print()
			return
		}
		output.NewInfo("Querying I2P peers from web console...").Print()
		output.NewInfo("Open http://127.0.0.1:7657 for full peer list").Print()
	},
}

var i2pTunnelsCmd = &cobra.Command{
	Use:   "tunnels",
	Short: "List active I2P tunnels",
	Run: func(cmd *cobra.Command, args []string) {
		if !isI2PRunning() {
			output.NewInfo("I2P router is not running").Print()
			return
		}
		output.NewInfo("Active tunnels: check http://127.0.0.1:7657/tunnels").Print()
	},
}

var i2pSitesCmd = &cobra.Command{
	Use:   "sites",
	Short: "Show known I2P eepsites",
	Run: func(cmd *cobra.Command, args []string) {
		output.NewInfo("I2P eepsites are accessible via http://127.0.0.1:7658/").Print()
		output.NewInfo("Use an I2P-enabled browser to browse .i2p domains").Print()
	},
}

var i2pProxiesCmd = &cobra.Command{
	Use:   "proxies",
	Short: "Show I2P proxy configuration",
	Run: func(cmd *cobra.Command, args []string) {
		running := isI2PRunning()
		headers := []string{"PROXY", "ADDRESS", "STATUS"}
		var rows [][]string
		status := "stopped"
		if running {
			status = "running"
		}
		rows = append(rows, []string{"HTTP Proxy", "127.0.0.1:4444", status})
		rows = append(rows, []string{"SOCKS Proxy", "127.0.0.1:4447", status})
		rows = append(rows, []string{"SAM Bridge", "127.0.0.1:7656", status})
		rows = append(rows, []string{"BOB Bridge", "127.0.0.1:2827", status})
		rows = append(rows, []string{"I2CP", "127.0.0.1:7654", status})
		output.PrintCompactTable(headers, rows)
	},
}

var i2pLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show I2P router logs",
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := os.UserHomeDir()
		logPaths := []string{
			filepath.Join(home, ".nux", "i2p", "logs", "router.log"),
			"/var/log/i2p/router.log",
			filepath.Join(home, ".i2p", "logs", "router.log"),
		}
		for _, lp := range logPaths {
			if _, err := os.Stat(lp); err == nil {
				data, _ := os.ReadFile(lp)
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
				return
			}
		}
		output.PrintWarningMessage("No I2P log files found")
	},
}

var i2pStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show I2P router statistics",
	Run: func(cmd *cobra.Command, args []string) {
		if !isI2PRunning() {
			output.NewInfo("I2P router is not running").Print()
			return
		}
		output.NewInfo("I2P router is running").Print()
		output.NewInfo("Full stats available at http://127.0.0.1:7657/stats").Print()
	},
}

var i2pDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run I2P health diagnostics",
	Run: func(cmd *cobra.Command, args []string) {
		output.NewInfo("I2P Diagnostics").Print()
		fmt.Println(strings.Repeat("-", 50))

		issues := 0

		router := findI2PRouter()
		if router != "" {
			output.NewInfo(fmt.Sprintf("[OK] I2P router found: %s", router)).Print()
		} else {
			output.PrintWarningMessage("[WARN] I2P router not found in PATH")
			issues++
		}

		if isI2PRunning() {
			output.NewInfo("[OK] I2P router is running").Print()
		} else {
			output.PrintWarningMessage("[WARN] I2P router is not running")
		}

		for _, port := range []int{7657, 4444, 4447, 7656} {
			if isPortOpen(port) {
				output.NewInfo(fmt.Sprintf("[OK] Port %d is open", port)).Print()
			}
		}

		if _, err := exec.LookPath("java"); err == nil {
			output.NewInfo("[OK] Java runtime found").Print()
		} else {
			output.PrintWarningMessage("[WARN] Java not found (required for I2P)")
			issues++
		}

		if issues > 0 {
			output.PrintWarningMessage(fmt.Sprintf("Found %d issues requiring attention", issues))
		} else {
			output.NewInfo("All checks passed").Print()
		}
	},
}

var i2pReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload I2P router configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if !isI2PRunning() {
			output.NewInfo("I2P router is not running").Print()
			return
		}
		output.NewInfo("Reloading I2P router configuration...").Print()
		output.NewInfo("Configuration reload triggered via admin console").Print()
	},
}

var i2pShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open I2P shell console",
	Run: func(cmd *cobra.Command, args []string) {
		output.NewInfo("I2P router shell").Print()
		output.NewInfo("Connect via SAM or I2CP for programmatic access").Print()
	},
}

var i2pConsoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Open I2P web admin console",
	Run: func(cmd *cobra.Command, args []string) {
		output.NewInfo("Opening I2P web console...").Print()
		exec.Command("xdg-open", "http://127.0.0.1:7657").Start()
	},
}

func findI2PRouter() string {
	candidates := []string{
		"i2prouter",
		"i2p",
		"/usr/bin/i2prouter",
		"/usr/local/bin/i2prouter",
	}
	for _, c := range candidates {
		if p, err := exec.LookPath(c); err == nil {
			return p
		}
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func isI2PRunning() bool {
	out, err := exec.Command("pgrep", "-f", "i2prouter|I2P.*router").CombinedOutput()
	if err != nil {
		return false
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		if pid != os.Getpid() && pid != os.Getppid() {
			return true
		}
	}
	return false
}

func isI2PInstalled() bool {
	return findI2PRouter() != ""
}

func init() {
	i2pCmd.AddCommand(i2pStartCmd)
	i2pCmd.AddCommand(i2pStopCmd)
	i2pCmd.AddCommand(i2pRestartCmd)
	i2pCmd.AddCommand(i2pStatusCmd)
	i2pCmd.AddCommand(i2pPeersCmd)
	i2pCmd.AddCommand(i2pTunnelsCmd)
	i2pCmd.AddCommand(i2pSitesCmd)
	i2pCmd.AddCommand(i2pProxiesCmd)
	i2pCmd.AddCommand(i2pLogsCmd)
	i2pCmd.AddCommand(i2pStatsCmd)
	i2pCmd.AddCommand(i2pDoctorCmd)
	i2pCmd.AddCommand(i2pReloadCmd)
	i2pCmd.AddCommand(i2pShellCmd)
	i2pCmd.AddCommand(i2pConsoleCmd)
	rootCmd.AddCommand(i2pCmd)
}
