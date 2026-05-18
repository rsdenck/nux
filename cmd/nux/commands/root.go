package commands

import (
	"fmt"
	"os"

	"github.com/rsdenck/nux/internal/core/logger"
	"github.com/rsdenck/nux/internal/output"
	"github.com/spf13/cobra"
)

var _ = fmt.Println
var _ = os.Stderr

var (
	// Global flags
	flagJSON    bool
	flagYAML    bool
	flagQuiet   bool
	flagDryRun  bool
	flagTimeout int
	flagVerbose bool
	flagNoColor bool
	flagLogFile string
)

var rootCmd = &cobra.Command{
	Use:   "nux",
	Short: "NUX — Linux CLI Manager",
	Long:  `Uma CLI de nível de produção para administração abrangente de sistemas Linux.`,
	Run: func(cmd *cobra.Command, args []string) {
		// ASCII Art com cor laranja (208)
		fmt.Print("\033[38;5;208m")
		fmt.Println(" ███╗   ██╗██╗   ██╗██╗  ██╗")
		fmt.Println(" ████╗  ██║██║   ██║╚██╗██╔╝")
		fmt.Println(" ██╔██╗ ██║██║   ██║ ╚███╔╝ ")
		fmt.Println(" ██║╚██╗██║██║   ██║ ██╔██╗ ")
		fmt.Println(" ██║ ╚████║╚██████╔╝██╔╝ ██╗")
		fmt.Println(" ╚═╝  ╚═══╝ ╚═════╝ ╚═╝  ╚═╝")
		fmt.Print("\033[0m")
		fmt.Println()
		fmt.Println("NUX — Linux Operations Platform")
		fmt.Println("┌──────────────────────────────────────────────────────────────────┐")
		fmt.Println("│ Manage Linux systems, automation, security, skills and AI agents │")
		fmt.Println("│ from one professional command-line interface.                    │")
		fmt.Println("└──────────────────────────────────────────────────────────────────┘")
		fmt.Println()
		fmt.Println("CORE MODULES")
		fmt.Println("┌──────────────────────────────────────────────────────────┐")
		fmt.Println("│  doctor        Run full health diagnostics               │")
		fmt.Println("│  system        Show OS, kernel, CPU, RAM, uptime         │")
		fmt.Println("│  disk          Manage disks, partitions and filesystems  │")
		fmt.Println("│  network       Network interfaces, routes and diagnostics│")
		fmt.Println("│  service       Manage system services                    │")
		fmt.Println("│  process       Process monitoring and control            │")
		fmt.Println("│  users         User and group management                 │")
		fmt.Println("│  firewall      Firewall management                       │")
		fmt.Println("└──────────────────────────────────────────────────────────┘")
		fmt.Println()
		fmt.Println("AUTOMATION & EXTENSIONS")
		fmt.Println("┌────────────────────────────────────────────────────────┐")
		fmt.Println("│  skill         Install and manage external CLI skills  │")
		fmt.Println("│  plugin        Legacy plugin compatibility             │")
		fmt.Println("│  bash          Execute controlled shell commands       │")
		fmt.Println("│  completion    Generate shell completion scripts       │")
		fmt.Println("└────────────────────────────────────────────────────────┘")
		fmt.Println()
	fmt.Println("SECURITY")
	fmt.Println("┌───────────────────────────────────────────────────────────┐")
	fmt.Println("│  audit         Security auditing and investigations       │")
	fmt.Println("│  wg            WireGuard VPN management                   │")
	fmt.Println("└───────────────────────────────────────────────────────────┘")
		fmt.Println()
		fmt.Println("CONTAINERS & REMOTE")
		fmt.Println("┌────────────────────────────────────────────────┐")
		fmt.Println("│  container     Docker / Podman management      │")
		fmt.Println("│  remote        SSH remote operations           │")
		fmt.Println("└────────────────────────────────────────────────┘")
		fmt.Println()
	fmt.Println("AI & P2P NETWORKS")
	fmt.Println("┌────────────────────────────────────────────────┐")
	fmt.Println("│  agent         Ollama Linux assistant          │")
	fmt.Println("│  znet          ZeroNet P2P network management  │")
	fmt.Println("│  i2p           I2P anonymous network manager   │")
	fmt.Println("└────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("SETUP")
	fmt.Println("┌────────────────────────────────────────────────┐")
	fmt.Println("│  onboard       First-time guided setup         │")
	fmt.Println("└────────────────────────────────────────────────┘")
		fmt.Println()
		fmt.Println("GLOBAL FLAGS")
		fmt.Println("┌────────────────────────────────────────────────┐")
		fmt.Println("│  --json        JSON output                     │")
		fmt.Println("│  --yaml        YAML output                     │")
		fmt.Println("│  --verbose     Verbose logs                    │")
		fmt.Println("│  --quiet       Quiet mode                      │")
		fmt.Println("│  --dry-run     Simulate actions                │")
		fmt.Println("│  --timeout     Command timeout                 │")
		fmt.Println("│  --no-color    Disable colors                  │")
		fmt.Println("│  -v            Show version                    │")
		fmt.Println("└────────────────────────────────────────────────┘")
		fmt.Println()
		fmt.Println("QUICK START")
		fmt.Println("┌────────────────────────────────────────────────┐")
		fmt.Println("│  nux onboard                                   │")
		fmt.Println("│  nux doctor                                    │")
		fmt.Println("│  nux service list                              │")
		fmt.Println("│  nux skill list                                │")
		fmt.Println("│  nux agent status                              │")
		fmt.Println("└────────────────────────────────────────────────┘")
		fmt.Println()
		fmt.Println("HELP")
		fmt.Println("┌────────────────────────────────────────────────┐")
		fmt.Println("│  nux <command> --help                          │")
		fmt.Println("└────────────────────────────────────────────────┘")
		fmt.Println()
		fmt.Println("VERSION")
		fmt.Println("┌────────────────────────────────────────────────┐")
		fmt.Println("│  NUX vdev                                      │")
		fmt.Println("└────────────────────────────────────────────────┘")
		fmt.Println()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		opts := logger.Options{
			Debug:    flagVerbose,
			UseColor: !flagNoColor,
			LogFile:  flagLogFile,
		}
		if err := logger.Init(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Falha ao inicializar logger: %v\n", err)
		}
		output.SetFormat(flagJSON, flagYAML)
	},
}

func Execute(version, commit, date string) {
	if version != "" {
		rootCmd.Version = version
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Saída em formato JSON")
	rootCmd.PersistentFlags().BoolVar(&flagYAML, "yaml", false, "Saída em formato YAML")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "Suprime a saída")
	rootCmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "Simula a execução sem fazer alterações")
	rootCmd.PersistentFlags().IntVar(&flagTimeout, "timeout", 30, "Tempo limite em segundos")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Ativa o log detalhado")
	rootCmd.PersistentFlags().StringVar(&flagLogFile, "log-file", "", "Caminho do arquivo de log (ativa o log de depuração para arquivo)")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Desativa saída colorida")
}
