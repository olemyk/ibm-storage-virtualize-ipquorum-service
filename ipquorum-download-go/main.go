package main

import (
	"fmt"
	"os"

	"github.com/olemyk/ipquorum-go/internal/errors"
	"github.com/olemyk/ipquorum-go/internal/utils"
	"github.com/olemyk/ipquorum-go/pkg/client"
	"github.com/olemyk/ipquorum-go/pkg/config"
	"github.com/olemyk/ipquorum-go/pkg/logger"
	"github.com/olemyk/ipquorum-go/pkg/password"
	"github.com/spf13/cobra"
)

var (
	// Config flags
	apiEndpoint    string
	username       string
	passwordStr    string
	passwordFile   string
	passwordPrompt bool
	outputFile     string
	debug          bool
	mkquorumapp    *bool
	download       *bool
	insecure       bool

	// mkquorumapp flags
	ip6           *bool
	nometadata    *bool
	partnersystem string
	partnerip6    *bool

	// Version information
	version = "1.0.0"
	commit  = "dev"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "ipquorum-download-go",
	Short: "IBM Storage Virtualize IP Quorum Download Tool",
	Long: `A high-performance Go tool to download IBM Storage Virtualize IP Quorum JAR files
and create Quorum Apps via REST API.

Features:
  - Single binary, no dependencies
  - Cross-platform (Linux, macOS, Windows)
  - Secure password handling
  - Smart retry logic with rate limiting
  - Fail-fast authentication
  - Professional logging`,
	RunE: run,
}

func init() {
	// General flags
	rootCmd.Flags().StringVar(&apiEndpoint, "api-endpoint", os.Getenv("API_ENDPOINT"), "API endpoint IP/hostname (REQUIRED)")
	rootCmd.Flags().StringVar(&username, "user", os.Getenv("VIRTUALIZE_USERNAME"), "Auth username")
	rootCmd.Flags().StringVar(&outputFile, "output", getEnvOrDefault("IPQ_OUTPUT_FILE", "ip_quorum.jar"), "Output jar filename")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.Flags().BoolVar(&insecure, "insecure", true, "Use insecure TLS (default)")
	rootCmd.Flags().Bool("secure", false, "Use strict TLS verification")

	// Operation flags
	mkquorumapp = rootCmd.Flags().Bool("mkquorumapp", true, "Enable mkquorumapp call")
	rootCmd.Flags().Bool("no-mkquorumapp", false, "Disable mkquorumapp call")
	download = rootCmd.Flags().Bool("download", true, "Enable jar download")
	rootCmd.Flags().Bool("no-download", false, "Disable jar download")

	// Password flags (mutually exclusive handled in code)
	rootCmd.Flags().StringVar(&passwordStr, "pass", os.Getenv("VIRTUALIZE_PASSWORD"), "Auth password (INSECURE - testing only)")
	rootCmd.Flags().StringVar(&passwordFile, "pass-file", "", "Read password from file (SECURE for automation)")
	rootCmd.Flags().BoolVar(&passwordPrompt, "pass-prompt", false, "Prompt for password interactively (MOST SECURE)")

	// mkquorumapp payload flags
	ip6 = rootCmd.Flags().Bool("ip6", false, "Set IPv6 flag")
	rootCmd.Flags().Bool("ip_6", false, "Set IPv6 flag (alias)")
	nometadata = rootCmd.Flags().Bool("nometadata", false, "Set nometadata flag")
	rootCmd.Flags().StringVar(&partnersystem, "partnersystem", "", "Set Partnersystem - Remote System in PBHA (MANDATORY if mkquorumapp enabled)")
	partnerip6 = rootCmd.Flags().Bool("partnerip6", false, "Set partner IPv6 flag")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ipquorum-download-go version %s\n", version)
			fmt.Printf("commit: %s\n", commit)
			fmt.Printf("built: %s\n", date)
		},
	})
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func run(cmd *cobra.Command, args []string) error {
	// Initialize logger
	log := logger.New(debug)

	// Handle --secure flag (overrides --insecure)
	if cmd.Flags().Changed("secure") {
		secure, _ := cmd.Flags().GetBool("secure")
		insecure = !secure
	}

	// Handle --no-mkquorumapp flag
	if cmd.Flags().Changed("no-mkquorumapp") {
		noMkquorumapp, _ := cmd.Flags().GetBool("no-mkquorumapp")
		if noMkquorumapp {
			*mkquorumapp = false
		}
	}

	// Handle --no-download flag
	if cmd.Flags().Changed("no-download") {
		noDownload, _ := cmd.Flags().GetBool("no-download")
		if noDownload {
			*download = false
		}
	}

	// Handle --ip_6 alias
	if cmd.Flags().Changed("ip_6") {
		ip6Val, _ := cmd.Flags().GetBool("ip_6")
		*ip6 = ip6Val
	}

	// Create configuration
	cfg := config.NewConfig()
	cfg.APIEndpoint = apiEndpoint
	cfg.Username = username
	cfg.OutputFile = outputFile
	cfg.Debug = debug
	cfg.VerifySSL = !insecure
	cfg.MkQuorumApp = *mkquorumapp
	cfg.Download = *download
	cfg.IP6 = *ip6
	cfg.NoMetadata = *nometadata
	cfg.PartnerSystem = partnersystem
	cfg.PartnerIP6 = *partnerip6

	// Handle password input
	var err error
	if passwordPrompt {
		cfg.Password, err = password.GetPassword("prompt", "", username)
	} else if passwordFile != "" {
		cfg.Password, err = password.GetPassword("file", passwordFile, username)
	} else if passwordStr != "" {
		cfg.Password = passwordStr
	} else {
		// No password provided, prompt interactively
		log.Info("No password provided, prompting interactively...")
		cfg.Password, err = password.GetPassword("prompt", "", username)
	}

	if err != nil {
		return err
	}

	// Debug output (with masked password)
	if debug {
		log.Debug("mkquorumapp=%t, download=%t, insecure=%t", cfg.MkQuorumApp, cfg.Download, insecure)
		log.Debug("ip6=%t, nometadata=%t, partnersystem='%s', partnerip6=%t",
			cfg.IP6, cfg.NoMetadata, cfg.PartnerSystem, cfg.PartnerIP6)
		log.Debug("user='%s', endpoint='%s'", cfg.Username, cfg.APIEndpoint)
		log.Debug("password='%s'", utils.MaskPassword(cfg.Password, 0))
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Create client
	c := client.NewClient(cfg, log)

	// Pre-flight check
	if err := c.PreflightCheck(); err != nil {
		return err
	}

	// Authenticate
	if err := c.Authenticate(); err != nil {
		return err
	}

	// Create quorum app (if enabled)
	if cfg.MkQuorumApp {
		if err := c.CreateQuorumApp(); err != nil {
			return err
		}
	} else {
		log.Info("Condition not met. Skipping the Create new IP-Quorum app call.")
	}

	// Download JAR (if enabled)
	if cfg.Download {
		if err := c.DownloadJAR(); err != nil {
			return err
		}
	} else {
		log.Info("Condition not met. Skipping the downloading of quorumapp.")
	}

	log.Info("Operation completed successfully")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Determine exit code based on error type
		exitCode := 1

		switch err.(type) {
		case *errors.ValidationError:
			exitCode = 2
		case *errors.NetworkError:
			exitCode = 3
		case *errors.AuthenticationError, *errors.APIError:
			exitCode = 1
		}

		os.Exit(exitCode)
	}
}

// Made with help from Bob
