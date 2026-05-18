package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rsdenck/nux/internal/output"
	"github.com/spf13/cobra"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var wgCmd = &cobra.Command{
	Use:   "wg",
	Short: "WireGuard VPN management",
	Long:  `Manage WireGuard interfaces, peers, and Cloudflare Warp via wg-quick, wgcf, and wgctrl.`,
}

var wgStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show WireGuard interface status from kernel",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := wgctrl.New()
		if err != nil {
			output.NewError(fmt.Sprintf("Failed to initialize wgctrl: %s", err.Error()), "WG_CTRL_ERROR").Print()
			return
		}
		defer client.Close()

		devices, err := client.Devices()
		if err != nil {
			output.NewError(fmt.Sprintf("Failed to list WireGuard devices: %s", err.Error()), "WG_LIST_ERROR").Print()
			return
		}

		if len(devices) == 0 {
			output.NewInfo("No WireGuard interfaces found").Print()
			return
		}

		for _, d := range devices {
			status := "active"
			color := "\033[32m"
			if len(d.Peers) == 0 {
				status = "inactive"
				color = "\033[33m"
			}

			handshake := ""
			if len(d.Peers) > 0 {
				latest := time.Time{}
				for _, p := range d.Peers {
					if p.LastHandshakeTime.After(latest) {
						latest = p.LastHandshakeTime
					}
				}
				if !latest.IsZero() {
					ago := time.Since(latest).Round(time.Second)
					handshake = fmt.Sprintf("%s ago", ago)
				}
			}

			fmt.Printf("%s%s %s\033[0m\n", color, d.Name, status)
			fmt.Printf("  Private key: %s\n", d.PrivateKey.String()[:16]+"...")
			fmt.Printf("  Public key:  %s\n", d.PublicKey.String()[:16]+"...")
			fmt.Printf("  Listen port: %d\n", d.ListenPort)
			fmt.Printf("  Peers:       %d\n", len(d.Peers))
			if handshake != "" {
				fmt.Printf("  Handshake:   %s\n", handshake)
			}
			if len(d.Peers) > 0 {
				fmt.Println("  Peers:")
				for _, p := range d.Peers {
					fmt.Printf("    - %s\n", p.PublicKey.String()[:16]+"...")
					if p.Endpoint != nil {
						fmt.Printf("      Endpoint: %s\n", p.Endpoint.String())
					}
					fmt.Printf("      Allowed IPs: %v\n", p.AllowedIPs)
					tx := p.TransmitBytes
					rx := p.ReceiveBytes
					fmt.Printf("      Transfer: %s sent, %s received\n", formatBytes(tx), formatBytes(rx))
				}
			}
			fmt.Println()
		}
	},
}

var wgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all WireGuard interfaces",
	Run: func(cmd *cobra.Command, args []string) {
		// Try kernel via wgctrl first
		client, err := wgctrl.New()
		if err == nil {
			defer client.Close()
			devices, err := client.Devices()
			if err == nil && len(devices) > 0 {
				headers := []string{"INTERFACE", "PUBLIC KEY", "PORT", "PEERS", "STATUS"}
				var rows [][]string
				for _, d := range devices {
					status := "active"
					if len(d.Peers) == 0 {
						status = "inactive"
					}
					pubKey := d.PublicKey.String()[:16] + "..."
					peers := fmt.Sprintf("%d", len(d.Peers))
					rows = append(rows, []string{d.Name, pubKey, fmt.Sprintf("%d", d.ListenPort), peers, status})
				}
				output.PrintCompactTable(headers, rows)
				return
			}
			_ = devices
		}

		// Fallback: try `wg show` CLI
		out, err := exec.Command("wg", "show", "interfaces").CombinedOutput()
		if err != nil {
			output.NewInfo("No WireGuard interfaces found").Print()
			return
		}
		interfaces := strings.Fields(strings.TrimSpace(string(out)))
		if len(interfaces) == 0 {
			output.NewInfo("No WireGuard interfaces found").Print()
			return
		}
		headers := []string{"INTERFACE"}
		var rows [][]string
		for _, iface := range interfaces {
			rows = append(rows, []string{iface})
		}
		output.PrintCompactTable(headers, rows)
	},
}

var wgConnectCmd = &cobra.Command{
	Use:   "connect [config-file]",
	Short: "Connect WireGuard interface via wg-quick",
	Long: `Connect a WireGuard interface using wg-quick.
If no config file is given, uses /etc/wireguard/*.conf.
Example: nux wg connect /etc/wireguard/wg0.conf`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Try auto-detecting configs in /etc/wireguard/
			configs, err := os.ReadDir("/etc/wireguard")
			if err != nil {
				output.NewError("No config specified and /etc/wireguard/ not found. Usage: nux wg connect <config.conf>", "WG_CONFIG_MISSING").Print()
				return
			}
			var confFiles []string
			for _, f := range configs {
				if strings.HasSuffix(f.Name(), ".conf") {
					confFiles = append(confFiles, f.Name())
				}
			}
			if len(confFiles) == 0 {
				output.NewError("No .conf files found in /etc/wireguard/. Specify a config file.", "WG_CONFIG_MISSING").Print()
				return
			}
			if len(confFiles) == 1 {
				args = []string{strings.TrimSuffix(confFiles[0], ".conf")}
			} else {
				output.NewError(fmt.Sprintf("Multiple configs found: %v. Specify one.", confFiles), "WG_MULTIPLE_CONFIGS").Print()
				return
			}
		}

		iface := strings.TrimSuffix(args[0], ".conf")
		output.NewInfo(fmt.Sprintf("Connecting WireGuard interface %s...", iface)).Print()

		out, err := exec.Command("wg-quick", "up", iface).CombinedOutput()
		if err != nil {
			output.NewError(fmt.Sprintf("Failed to connect: %s", strings.TrimSpace(string(out))), "WG_CONNECT_ERROR").Print()
			return
		}
		output.NewSuccess(map[string]interface{}{
			"action":    "connect",
			"interface": iface,
			"status":    "connected",
		}).Print()
	},
}

var wgDisconnectCmd = &cobra.Command{
	Use:   "disconnect [interface]",
	Short: "Disconnect WireGuard interface via wg-quick",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		iface := "wg0"
		if len(args) > 0 {
			iface = args[0]
		}

		output.NewInfo(fmt.Sprintf("Disconnecting WireGuard interface %s...", iface)).Print()

		out, err := exec.Command("wg-quick", "down", iface).CombinedOutput()
		if err != nil {
			output.NewError(fmt.Sprintf("Failed to disconnect: %s", strings.TrimSpace(string(out))), "WG_DISCONNECT_ERROR").Print()
			return
		}
		output.NewSuccess(map[string]interface{}{
			"action":    "disconnect",
			"interface": iface,
			"status":    "disconnected",
		}).Print()
	},
}

var wgWarpCmd = &cobra.Command{
	Use:   "warp",
	Short: "Cloudflare Warp management via wgcf",
	Long: `Manage Cloudflare Warp using wgcf.
Generate config, register, connect and disconnect from Warp.
Requires wgcf binary (https://github.com/ViRb3/wgcf)`,
}

var wgWarpGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Cloudflare Warp config with wgcf",
	Run: func(cmd *cobra.Command, args []string) {
		output.NewInfo("Generating Cloudflare Warp configuration...").Print()

		if _, err := exec.LookPath("wgcf"); err != nil {
			output.NewError("wgcf not found. Install it first: https://github.com/ViRb3/wgcf/releases", "WGCF_MISSING").Print()
			return
		}

		out, err := exec.Command("wgcf", "generate").CombinedOutput()
		if err != nil {
			output.NewError(fmt.Sprintf("Failed to generate Warp config: %s", strings.TrimSpace(string(out))), "WGCF_GENERATE_ERROR").Print()
			return
		}
		output.NewSuccess(map[string]interface{}{
			"action": "generate",
			"status": "config_generated",
		}).Print()
	},
}

var wgWarpRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register with Cloudflare Warp via wgcf",
	Run: func(cmd *cobra.Command, args []string) {
		output.NewInfo("Registering with Cloudflare Warp...").Print()

		if _, err := exec.LookPath("wgcf"); err != nil {
			output.NewError("wgcf not found. Install it first: https://github.com/ViRb3/wgcf/releases", "WGCF_MISSING").Print()
			return
		}

		out, err := exec.Command("wgcf", "register", "--accept-tos").CombinedOutput()
		if err != nil {
			output.NewError(fmt.Sprintf("Failed to register: %s", strings.TrimSpace(string(out))), "WGCF_REGISTER_ERROR").Print()
			return
		}
		output.NewSuccess(map[string]interface{}{
			"action": "register",
			"status": "registered",
		}).Print()
	},
}

var wgWarpConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect Cloudflare Warp via wgcf",
	Run: func(cmd *cobra.Command, args []string) {
		output.NewInfo("Connecting to Cloudflare Warp...").Print()

		if _, err := exec.LookPath("wgcf"); err != nil {
			output.NewError("wgcf not found. Install it first: https://github.com/ViRb3/wgcf/releases", "WGCF_MISSING").Print()
			return
		}

		out, err := exec.Command("wgcf", "connect").CombinedOutput()
		if err != nil {
			output.NewError(fmt.Sprintf("Failed to connect to Warp: %s", strings.TrimSpace(string(out))), "WGCF_CONNECT_ERROR").Print()
			return
		}
		output.NewSuccess(map[string]interface{}{
			"action": "warp_connect",
			"status": "connected",
		}).Print()
	},
}

var wgWarpDisconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from Cloudflare Warp",
	Run: func(cmd *cobra.Command, args []string) {
		output.NewInfo("Disconnecting from Cloudflare Warp...").Print()

		out, err := exec.Command("wg-quick", "down", "wgcf").CombinedOutput()
		if err != nil {
			out2, err2 := exec.Command("wgcf", "disconnect").CombinedOutput()
			if err2 != nil {
				output.NewError(fmt.Sprintf("Failed to disconnect Warp: %s", strings.TrimSpace(string(out2))), "WGCF_DISCONNECT_ERROR").Print()
				return
			}
			_ = out2
		}
		_ = out
		output.NewSuccess(map[string]interface{}{
			"action": "warp_disconnect",
			"status": "disconnected",
		}).Print()
	},
}

var wgWarpStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Cloudflare Warp connection status",
	Run: func(cmd *cobra.Command, args []string) {
		// Check via wgctrl if wgcf interface exists
		client, err := wgctrl.New()
		if err == nil {
			defer client.Close()
			devices, _ := client.Devices()
			for _, d := range devices {
				if d.Name == "wgcf" {
					color := "\033[32m"
					status := "connected"
					if len(d.Peers) == 0 {
						color = "\033[33m"
						status = "inactive"
					}
					fmt.Printf("%sCloudflare Warp [%s]\033[0m\n", color, status)
					fmt.Printf("  Public key:  %s\n", d.PublicKey.String()[:16]+"...")
					fmt.Printf("  Peers:       %d\n", len(d.Peers))
					return
				}
			}
		}
		// Fallback: try resolving warp
		out, err := exec.Command("wg", "show", "wgcf").CombinedOutput()
		if err != nil {
			output.NewInfo("Cloudflare Warp is not connected").Print()
			return
		}
		fmt.Printf("Cloudflare Warp [connected]\n%s\n", string(out))
	},
}

var wgInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install WireGuard tools (wg, wg-quick, wgcf)",
	Run: func(cmd *cobra.Command, args []string) {
		output.NewInfo("Installing WireGuard tools...").Print()

		out, err := exec.Command("cat", "/etc/os-release").CombinedOutput()
		if err != nil {
			output.NewError("Failed to detect OS", "WG_OS_DETECT").Print()
			return
		}
		osRelease := strings.ToLower(string(out))

		var installCmd *exec.Cmd
		switch {
		case strings.Contains(osRelease, "rocky") || strings.Contains(osRelease, "rhel") || strings.Contains(osRelease, "centos"):
			installCmd = exec.Command("sh", "-c", "dnf install -y wireguard-tools && curl -fsSL https://github.com/ViRb3/wgcf/releases/latest/download/wgcf_2.2.22_linux_amd64 -o /usr/local/bin/wgcf && chmod +x /usr/local/bin/wgcf")
		case strings.Contains(osRelease, "fedora"):
			installCmd = exec.Command("sh", "-c", "dnf install -y wireguard-tools && curl -fsSL https://github.com/ViRb3/wgcf/releases/latest/download/wgcf_2.2.22_linux_amd64 -o /usr/local/bin/wgcf && chmod +x /usr/local/bin/wgcf")
		case strings.Contains(osRelease, "ubuntu") || strings.Contains(osRelease, "debian"):
			installCmd = exec.Command("sh", "-c", "apt-get install -y wireguard-tools && curl -fsSL https://github.com/ViRb3/wgcf/releases/latest/download/wgcf_2.2.22_linux_amd64 -o /usr/local/bin/wgcf && chmod +x /usr/local/bin/wgcf")
		case strings.Contains(osRelease, "arch"):
			installCmd = exec.Command("sh", "-c", "pacman -S --noconfirm wireguard-tools && curl -fsSL https://github.com/ViRb3/wgcf/releases/latest/download/wgcf_2.2.22_linux_amd64 -o /usr/local/bin/wgcf && chmod +x /usr/local/bin/wgcf")
		default:
			installCmd = exec.Command("sh", "-c", "curl -fsSL https://github.com/ViRb3/wgcf/releases/latest/download/wgcf_2.2.22_linux_amd64 -o /usr/local/bin/wgcf && chmod +x /usr/local/bin/wgcf")
		}

		installOut, err := installCmd.CombinedOutput()
		if err != nil {
			output.NewError(fmt.Sprintf("Installation failed: %s", strings.TrimSpace(string(installOut))), "WG_INSTALL_ERROR").Print()
			return
		}
		output.NewSuccess(map[string]interface{}{
			"action": "install",
			"status": "wireguard_tools_installed",
		}).Print()
	},
}

var wgQuickStatusCmd = &cobra.Command{
	Use:   "quick-status",
	Short: "Show wg-quick managed interface status",
	Run: func(cmd *cobra.Command, args []string) {
		// Check systemd for wg-quick services
		out, err := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager").CombinedOutput()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			headers := []string{"SERVICE", "STATUS"}
			var rows [][]string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "wg-quick") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						serviceName := parts[0]
						status := parts[2]
						rows = append(rows, []string{serviceName, status})
					}
				}
			}
			if len(rows) > 0 {
				output.PrintCompactTable(headers, rows)
			} else {
				output.NewInfo("No wg-quick services found").Print()
			}
			return
		}
		output.NewInfo("systemctl not available").Print()
	},
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

var wgShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show raw WireGuard interface configuration",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		wgArgs := []string{"show"}
		if len(args) > 0 {
			wgArgs = append(wgArgs, args[0])
		}
		out, err := exec.Command("wg", wgArgs...).CombinedOutput()
		if err != nil {
			output.NewError(fmt.Sprintf("wg show failed: %s", strings.TrimSpace(string(out))), "WG_SHOW_ERROR").Print()
			return
		}
		fmt.Print(string(out))
	},
}

var wgGenkeyCmd = &cobra.Command{
	Use:   "genkey",
	Short: "Generate WireGuard keypair",
	Run: func(cmd *cobra.Command, args []string) {
		privateKey, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			output.NewError(fmt.Sprintf("Failed to generate key: %s", err.Error()), "WG_KEY_ERROR").Print()
			return
		}
		publicKey := privateKey.PublicKey()
		fmt.Printf("Private key: %s\n", privateKey.String())
		fmt.Printf("Public key:  %s\n", publicKey.String())
	},
}

var wgGenpskCmd = &cobra.Command{
	Use:   "genpsk",
	Short: "Generate WireGuard pre-shared key",
	Run: func(cmd *cobra.Command, args []string) {
		key, err := wgtypes.GenerateKey()
		if err != nil {
			output.NewError(fmt.Sprintf("Failed to generate PSK: %s", err.Error()), "WG_PSK_ERROR").Print()
			return
		}
		fmt.Printf("Pre-shared key: %s\n", key.String())
	},
}

func init() {
	wgCmd.AddCommand(wgStatusCmd)
	wgCmd.AddCommand(wgListCmd)
	wgCmd.AddCommand(wgConnectCmd)
	wgCmd.AddCommand(wgDisconnectCmd)
	wgCmd.AddCommand(wgShowCmd)
	wgCmd.AddCommand(wgGenkeyCmd)
	wgCmd.AddCommand(wgGenpskCmd)
	wgCmd.AddCommand(wgInstallCmd)
	wgCmd.AddCommand(wgQuickStatusCmd)

	wgWarpCmd.AddCommand(wgWarpGenerateCmd)
	wgWarpCmd.AddCommand(wgWarpRegisterCmd)
	wgWarpCmd.AddCommand(wgWarpConnectCmd)
	wgWarpCmd.AddCommand(wgWarpDisconnectCmd)
	wgWarpCmd.AddCommand(wgWarpStatusCmd)
	wgCmd.AddCommand(wgWarpCmd)

	rootCmd.AddCommand(wgCmd)
}
