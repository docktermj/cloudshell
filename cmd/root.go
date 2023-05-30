/*
Reference: https://github.com/zephinzer/cloudshell/blob/master/cmd/cloudshell/main.go
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docktermj/cloudshell/xtermserver"
	"github.com/senzing/senzing-tools/constant"
	"github.com/senzing/senzing-tools/helper"
	"github.com/senzing/senzing-tools/option"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultCommand              string = "/bin/bash"
	defaultConnectionErrorLimit int    = 10
	defaultKeepalivePingTimeout int    = 20
	defaultMaxBufferSizeBytes   int    = 512
	defaultPathLiveness         string = "/liveness"
	defaultPathMetrics          string = "/metrics"
	defaultPathReadiness        string = "/readiness"
	defaultPathXtermjs          string = "/xterm.js"
	defaultPort                 int    = 8261
	defaultServerAddr           string = "0.0.0.0"
	defaultWorkingDir           string = "."
	envarAllowedHostnames       string = "SENZING_TOOLS_ALLOWED_HOSTNAMES"
	envarArguments              string = "SENZING_TOOLS_ARGUMENTS"
	envarCommand                string = "SENZING_TOOLS_COMMAND"
	envarConnectionErrorLimit   string = "SENZING_TOOLS_CONNECTION_ERROR_LIMIT"
	envarKeepalivePingTimeout   string = "SENZING_TOOLS_KEEPALIVE_PING_TIMEOUT"
	envarMaxBufferSizeBytes     string = "SENZING_TOOLS_MAX_BUFFER_SIZE_BYTES"
	envarPathLiveness           string = "SENZING_TOOLS_PATH_LIVENESS"
	envarPathMetrics            string = "SENZING_TOOLS_PATH_METRICS"
	envarPathReadiness          string = "SENZING_TOOLS_PATH_READINESS"
	envarPathXtermjs            string = "SENZING_TOOLS_PATH_XTERMJS"
	envarPort                   string = "SENZING_TOOLS_SERVER_PORT"
	envarServerAddr             string = "SENZING_TOOLS_SERVER_ADDR"
	envarWorkingDir             string = "SENZING_TOOLS_WORKDIR"
	optionAllowedHostnames      string = "allowed-hostnames"
	optionArguments             string = "arguments"
	optionCommand               string = "command"
	optionConnectionErrorLimit  string = "connection-error-limit"
	optionKeepalivePingTimeout  string = "keepalive-ping-timeout"
	optionMaxBufferSizeBytes    string = "max-buffer-size-bytes"
	optionPathLiveness          string = "path-liveness"
	optionPathMetrics           string = "path-metrics"
	optionPathReadiness         string = "path-readiness"
	optionPathXtermjs           string = "path-xtermjs"
	optionPort                  string = "server-port"
	optionServerAddr            string = "server-addr"
	optionWorkingDir            string = "workdir"
	Short                       string = "view-xterm short description"
	Use                         string = "view-xterm"
	Long                        string = `
view-xterm long description.
	`
)

var (
	defaultAllowedHostnames []string = []string{"localhost"}
	defaultArguments        []string
)

// ----------------------------------------------------------------------------
// Private functions
// ----------------------------------------------------------------------------

// Since init() is always invoked, define command line parameters.
func init() {
	RootCmd.Flags().Int(optionConnectionErrorLimit, defaultConnectionErrorLimit, fmt.Sprintf("Connection re-attempts before terminating [%s]", envarConnectionErrorLimit))
	RootCmd.Flags().Int(optionKeepalivePingTimeout, defaultKeepalivePingTimeout, fmt.Sprintf("Maximum allowable seconds between a ping message and its response [%s]", envarKeepalivePingTimeout))
	RootCmd.Flags().Int(optionMaxBufferSizeBytes, defaultMaxBufferSizeBytes, fmt.Sprintf("Maximum length of terminal input [%s]", envarMaxBufferSizeBytes))
	RootCmd.Flags().Int(optionPort, defaultPort, fmt.Sprintf("Port the server listens on [%s]", envarPort))
	RootCmd.Flags().String(optionCommand, defaultCommand, fmt.Sprintf("Path of shell command [%s]", envarCommand))
	RootCmd.Flags().String(optionPathLiveness, defaultPathLiveness, fmt.Sprintf("URL for liveness probe [%s]", envarPathLiveness))
	RootCmd.Flags().String(optionPathMetrics, defaultPathMetrics, fmt.Sprintf("URL for prometheus metrics [%s]", envarPathMetrics))
	RootCmd.Flags().String(optionPathReadiness, defaultPathReadiness, fmt.Sprintf("URL for readiness probe [%s]", envarPathReadiness))
	RootCmd.Flags().String(optionPathXtermjs, defaultPathXtermjs, fmt.Sprintf("URL for xterm.js to attach [%s]", envarPathXtermjs))
	RootCmd.Flags().String(optionServerAddr, defaultServerAddr, fmt.Sprintf("IP interface server listens on [%s]", envarServerAddr))
	RootCmd.Flags().String(optionWorkingDir, defaultWorkingDir, fmt.Sprintf("Working directory [%s]", envarWorkingDir))
	RootCmd.Flags().StringSlice(optionAllowedHostnames, defaultAllowedHostnames, fmt.Sprintf("Comma-delimited list of hostnames permitted to connect to the websocket [%s]", envarAllowedHostnames))
	RootCmd.Flags().StringSlice(optionArguments, defaultArguments, fmt.Sprintf("Comma-delimited list of arguments passed to the terminal command prompt [%s]", envarArguments))

}

// If a configuration file is present, load it.
func loadConfigurationFile(cobraCommand *cobra.Command) {
	configuration := ""
	configFlag := cobraCommand.Flags().Lookup(option.Configuration)
	if configFlag != nil {
		configuration = configFlag.Value.String()
	}
	if configuration != "" { // Use configuration file specified as a command line option.
		viper.SetConfigFile(configuration)
	} else { // Search for a configuration file.

		// Determine home directory.

		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Specify configuration file name.

		viper.SetConfigName("view-xterm")
		viper.SetConfigType("yaml")

		// Define search path order.

		viper.AddConfigPath(home + "/.senzing-tools")
		viper.AddConfigPath(home)
		viper.AddConfigPath("/etc/senzing-tools")
	}

	// If a config file is found, read it in.

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Applying configuration file:", viper.ConfigFileUsed())
	}
}

// Configure Viper with user-specified options.
func loadOptions(cobraCommand *cobra.Command) {
	var err error = nil
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix(constant.SetEnvPrefix)

	// Ints

	intOptions := map[string]int{
		optionConnectionErrorLimit: defaultConnectionErrorLimit,
		optionKeepalivePingTimeout: defaultKeepalivePingTimeout,
		optionMaxBufferSizeBytes:   defaultMaxBufferSizeBytes,
		optionPort:                 defaultPort,
	}
	for optionKey, optionValue := range intOptions {
		viper.SetDefault(optionKey, optionValue)
		err = viper.BindPFlag(optionKey, cobraCommand.Flags().Lookup(optionKey))
		if err != nil {
			panic(err)
		}
	}

	// Strings

	stringOptions := map[string]string{
		optionCommand:       defaultCommand,
		optionPathLiveness:  defaultPathLiveness,
		optionPathMetrics:   defaultPathMetrics,
		optionPathReadiness: defaultPathReadiness,
		optionPathXtermjs:   defaultPathXtermjs,
		optionServerAddr:    defaultServerAddr,
		optionWorkingDir:    defaultWorkingDir,
	}
	for optionKey, optionValue := range stringOptions {
		viper.SetDefault(optionKey, optionValue)
		err = viper.BindPFlag(optionKey, cobraCommand.Flags().Lookup(optionKey))
		if err != nil {
			panic(err)
		}
	}

	// StringSlice

	stringSliceOptions := map[string][]string{
		optionAllowedHostnames: defaultAllowedHostnames,
		optionArguments:        defaultArguments,
	}
	for optionKey, optionValue := range stringSliceOptions {
		viper.SetDefault(optionKey, optionValue)
		err = viper.BindPFlag(optionKey, cobraCommand.Flags().Lookup(optionKey))
		if err != nil {
			panic(err)
		}
	}

}

// ----------------------------------------------------------------------------
// Public functions
// ----------------------------------------------------------------------------

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// Used in construction of cobra.Command
func PreRun(cobraCommand *cobra.Command, args []string) {
	loadConfigurationFile(cobraCommand)
	loadOptions(cobraCommand)
	cobraCommand.SetVersionTemplate(constant.VersionTemplate)
}

// Used in construction of cobra.Command
func RunE(_ *cobra.Command, _ []string) error {
	var err error = nil
	ctx := context.TODO()

	// Create object and Serve.

	xtermServer := &xtermserver.XtermServerImpl{
		AllowedHostnames:     viper.GetStringSlice(optionAllowedHostnames),
		Arguments:            viper.GetStringSlice(optionArguments),
		Command:              viper.GetString(optionCommand),
		ConnectionErrorLimit: viper.GetInt(optionConnectionErrorLimit),
		KeepalivePingTimeout: viper.GetInt(optionKeepalivePingTimeout),
		MaxBufferSizeBytes:   viper.GetInt(optionMaxBufferSizeBytes),
		PathLiveness:         viper.GetString(optionPathLiveness),
		PathMetrics:          viper.GetString(optionPathMetrics),
		PathReadiness:        viper.GetString(optionPathReadiness),
		PathXtermjs:          viper.GetString(optionPathXtermjs),
		ServerAddr:           viper.GetString(optionServerAddr),
		Port:                 viper.GetInt(optionPort),
		WorkingDir:           viper.GetString(optionWorkingDir),
	}
	err = xtermServer.Serve(ctx)
	return err
}

// Used in construction of cobra.Command
func Version() string {
	return helper.MakeVersion(githubVersion, githubIteration)
}

// ----------------------------------------------------------------------------
// Command
// ----------------------------------------------------------------------------

// RootCmd represents the command.
var RootCmd = &cobra.Command{
	Use:     Use,
	Short:   Short,
	Long:    Long,
	PreRun:  PreRun,
	RunE:    RunE,
	Version: Version(),
}
