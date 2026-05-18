package commands

import (
	"fmt"
	"strings"

	"github.com/rsdenck/nux/internal/core"
	"github.com/rsdenck/nux/internal/output"
	"github.com/spf13/cobra"
)

var serviceExecutor core.Executor = &core.RealExecutor{}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service management",
	Long:  `Manage system services (systemd, openrc, sysvinit, runit).`,
}

// Top-level convenience commands (short aliases)
var statusCmd = &cobra.Command{
	Use:   "status <service>",
	Short: "Show service status (alias for 'service status')",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceStatusCmd.Run(cmd, args)
	},
}

var startCmd = &cobra.Command{
	Use:   "start <service>",
	Short: "Start a service (alias for 'service start')",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceStartCmd.Run(cmd, args)
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop <service>",
	Short: "Stop a service (alias for 'service stop')",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceStopCmd.Run(cmd, args)
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart <service>",
	Short: "Restart a service (alias for 'service restart')",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceRestartCmd.Run(cmd, args)
	},
}

type ServiceInfo struct {
	Name    string
	State   string
	Enabled string
	Pid     string
	Ports   string
}

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List services",
	Run: func(cmd *cobra.Command, args []string) {
		out, err := serviceExecutor.CombinedOutput("systemctl", "list-units", "--type=service", "--all", "--no-pager")
		
		if err != nil {
			output.NewError(fmt.Sprintf("failed to list services: %s", strings.TrimSpace(out)), "SERVICE_LIST_ERROR").Print()
			return
		}

		services := parseServiceOutput(out)

		items := []map[string]interface{}{}
		for _, s := range services {
			items = append(items, map[string]interface{}{
				"name":    s.Name,
				"state":   s.State,
				"enabled": s.Enabled,
				"pid":     s.Pid,
				"ports":   s.Ports,
			})
		}

		output.NewList(items, len(items)).WithMessage("Service list").Print()
	},
}

func parseServiceOutput(out string) []ServiceInfo {
	services := []ServiceInfo{}
	lines := strings.Split(out, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "UNIT") || strings.HasPrefix(line, "●") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		name := strings.TrimSuffix(fields[0], ".service")
		state := fields[2]
		
		if state == "active" {
			state = "active"
		} else if state == "inactive" || state == "dead" {
			state = "inactive"
		}

		enabled := "no"
		if state == "active" {
			enabled = "yes"
		}

		pid := "-"
		pidOut, _ := serviceExecutor.CombinedOutput("systemctl", "show", name, "--property=MainPID", "--value")
		pidStr := strings.TrimSpace(pidOut)
		if pidStr != "" && pidStr != "0" {
			pid = pidStr
		}

		services = append(services, ServiceInfo{
			Name:    name,
			State:   state,
			Enabled: enabled,
			Pid:     pid,
			Ports:   "-",
		})
	}

	return services
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status <service>",
	Short: "Show service status",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		service := core.SanitizeInput(args[0])

		// Get status output
		out, _ := serviceExecutor.CombinedOutput("systemctl", "status", service)
		lines := strings.Split(strings.TrimSpace(out), "\n")

		// Color codes
		green := "\033[92m"
		red := "\033[91m"
		orange := "\033[93m"
		reset := "\033[0m"

		// Parse key service info
		keyInfo := map[string]string{}
		logs := []string{}
		inLogs := false
		isActive := false
		hasProblem := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}

			// Parse key metadata
			if strings.HasPrefix(trimmed, "Active:") {
				active := strings.TrimPrefix(trimmed, "Active: ")
				keyInfo["Active"] = active
				if strings.Contains(active, "active (running)") {
					isActive = true
				} else if strings.Contains(active, "failed") || strings.Contains(active, "error") {
					hasProblem = true
				}
			} else if strings.HasPrefix(trimmed, "Loaded:") {
				keyInfo["Loaded"] = strings.TrimPrefix(trimmed, "Loaded: ")
			} else if strings.HasPrefix(trimmed, "Main PID:") {
				keyInfo["Main PID"] = strings.TrimPrefix(trimmed, "Main PID: ")
			} else if strings.HasPrefix(trimmed, "Status:") {
				keyInfo["Status"] = strings.TrimPrefix(trimmed, "Status: ")
			} else if strings.HasPrefix(trimmed, "Tasks:") {
				keyInfo["Tasks"] = strings.TrimPrefix(trimmed, "Tasks: ")
			} else if strings.HasPrefix(trimmed, "Memory:") {
				keyInfo["Memory"] = strings.TrimPrefix(trimmed, "Memory: ")
			} else if strings.HasPrefix(trimmed, "CPU:") {
				keyInfo["CPU"] = strings.TrimPrefix(trimmed, "CPU: ")
			} else if strings.Contains(trimmed, "systemd[1]:") || strings.Contains(trimmed, "Started ") {
				inLogs = true
			}

			// Collect log lines
			if inLogs && (strings.Contains(trimmed, "systemd[1]:") || strings.Contains(trimmed, "httpd[") || strings.Contains(trimmed, "Started ")) {
				logs = append(logs, trimmed)
			}
		}

		// Determine status color
		statusColor := green
		statusText := "RUNNING"
		if hasProblem {
			statusColor = red
			statusText = "PROBLEM"
		} else if !isActive {
			statusColor = orange
			statusText = "STOPPED"
		}

	// Calculate table width based on content (NO width capping)
	// Header widths
	keyWidth := len("KEY") + 2
	valueWidth := len("VALUE") + 2
	
	// Adjust based on key info content
	order := []string{"Active", "Loaded", "Main PID", "Status", "Tasks", "Memory", "CPU"}
	for _, k := range order {
		if v, ok := keyInfo[k]; ok {
			// Key width
			if len(k)+2 > keyWidth {
				keyWidth = len(k) + 2
			}
			// Value width (use raw length, no truncation)
			if len(v)+2 > valueWidth {
				valueWidth = len(v) + 2
			}
		}
	}
	
	// Total table width: ┌ + keyWidth + ┬ + valueWidth + ┐ = keyWidth + 1 + valueWidth + 2
	tableWidth := keyWidth + 1 + valueWidth + 2

	// Print concise colored output
	fmt.Printf("\n● Service: %s [%s%s%s]\n", service, statusColor, statusText, reset)
	// Use tableWidth for separator to match table borders exactly
	fmt.Println(strings.Repeat("─", tableWidth))

		// Print key info as compact table
		if len(keyInfo) > 0 {
			headers := []string{"KEY", "VALUE"}
			rows := [][]string{}
			order := []string{"Active", "Loaded", "Main PID", "Status", "Tasks", "Memory", "CPU"}
			for _, k := range order {
				if v, ok := keyInfo[k]; ok {
					// Color the Active field
					if k == "Active" {
						if isActive {
							v = green + v + reset
						} else if hasProblem {
							v = red + v + reset
						} else {
							v = orange + v + reset
						}
					}
					rows = append(rows, []string{k, v})
				}
			}
			output.PrintCompactTable(headers, rows)
			fmt.Println()
		}

		// Print last 3 log lines
		if len(logs) > 0 {
			fmt.Println("Recent Logs (last 3):")
			fmt.Println(strings.Repeat("─", tableWidth))
			start := len(logs) - 3
			if start < 0 {
				start = 0
			}
			for _, log := range logs[start:] {
				fmt.Println(log)
			}
			fmt.Println()
		}
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start <service>",
	Short: "Start a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		service := core.SanitizeInput(args[0])
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		green := "\033[92m"
		reset := "\033[0m"

		if dryRun {
			output.NewInfo(map[string]interface{}{
				"service": service,
				"action":  "start",
				"dry_run": true,
				"command": fmt.Sprintf("systemctl start %s", service),
			}).Print()
			return
		}

		_, err := serviceExecutor.CombinedOutput("systemctl", "start", service)
		
		if err != nil {
			output.NewError(fmt.Sprintf("failed to start service: %s", err.Error()), "SERVICE_START_ERROR").Print()
			return
		}

		fmt.Printf("%s✔ Service %s started successfully%s\n", green, service, reset)
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop <service>",
	Short: "Stop a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		service := core.SanitizeInput(args[0])
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		orange := "\033[93m"
		reset := "\033[0m"

		if dryRun {
			output.NewInfo(map[string]interface{}{
				"service": service,
				"action":  "stop",
				"dry_run": true,
				"command": fmt.Sprintf("systemctl stop %s", service),
			}).Print()
			return
		}

		_, err := serviceExecutor.CombinedOutput("systemctl", "stop", service)
		
		if err != nil {
			output.NewError(fmt.Sprintf("failed to stop service: %s", err.Error()), "SERVICE_STOP_ERROR").Print()
			return
		}

		fmt.Printf("%s■ Service %s stopped%s\n", orange, service, reset)
	},
}

var serviceEnableCmd = &cobra.Command{
	Use:   "enable <service>",
	Short: "Enable service at boot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		service := core.SanitizeInput(args[0])
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			output.NewInfo(map[string]interface{}{
				"service": service,
				"action":  "enable",
				"dry_run": true,
				"command": fmt.Sprintf("systemctl enable %s", service),
			}).Print()
			return
		}

		_, err := serviceExecutor.CombinedOutput("systemctl", "enable", service)
		
		if err != nil {
			output.NewError(fmt.Sprintf("failed to enable service: %s", err.Error()), "SERVICE_ENABLE_ERROR").Print()
			return
		}

		output.NewSuccess(map[string]interface{}{
			"service": service,
			"action":  "enable",
			"status":  "enabled",
		}).Print()
	},
}

var serviceDisableCmd = &cobra.Command{
	Use:   "disable <service>",
	Short: "Disable service at boot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		service := core.SanitizeInput(args[0])
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			output.NewInfo(map[string]interface{}{
				"service": service,
				"action":  "disable",
				"dry_run": true,
				"command": fmt.Sprintf("systemctl disable %s", service),
			}).Print()
			return
		}

		_, err := serviceExecutor.CombinedOutput("systemctl", "disable", service)
		
		if err != nil {
			output.NewError(fmt.Sprintf("failed to disable service: %s", err.Error()), "SERVICE_DISABLE_ERROR").Print()
			return
		}

		output.NewSuccess(map[string]interface{}{
			"service": service,
			"action":  "disable",
			"status":  "disabled",
		}).Print()
	},
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart <service>",
	Short: "Restart a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		service := core.SanitizeInput(args[0])
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		green := "\033[92m"
		reset := "\033[0m"

		if dryRun {
			output.NewInfo(map[string]interface{}{
				"service": service,
				"action":  "restart",
				"dry_run": true,
				"command": fmt.Sprintf("systemctl restart %s", service),
			}).Print()
			return
		}

		_, err := serviceExecutor.CombinedOutput("systemctl", "restart", service)
		
		if err != nil {
			output.NewError(fmt.Sprintf("failed to restart service: %s", err.Error()), "SERVICE_RESTART_ERROR").Print()
			return
		}

		fmt.Printf("%s↻ Service %s restarted successfully%s\n", green, service, reset)
	},
}

func init() {
	serviceListCmd.Flags().Bool("json", false, "Output in JSON format")
	serviceCmd.AddCommand(serviceListCmd)
	serviceCmd.AddCommand(serviceStatusCmd)
	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceStopCmd)
	serviceCmd.AddCommand(serviceRestartCmd)
	serviceCmd.AddCommand(serviceEnableCmd)
	serviceCmd.AddCommand(serviceDisableCmd)
	
	serviceStartCmd.Flags().Bool("dry-run", false, "Simulate command")
	serviceStopCmd.Flags().Bool("dry-run", false, "Simulate command")
	serviceRestartCmd.Flags().Bool("dry-run", false, "Simulate command")
	serviceEnableCmd.Flags().Bool("dry-run", false, "Simulate command")
	serviceDisableCmd.Flags().Bool("dry-run", false, "Simulate command")
	
	rootCmd.AddCommand(serviceCmd)
	
	// Register top-level convenience commands
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
}
