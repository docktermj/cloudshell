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
	defaultServerAddress             string = "0.0.0.0"
	defaultServerPort                int    = 8261
	defaultXtermCommand              string = "/bin/bash"
	defaultXtermConnectionErrorLimit int    = 10
	defaultXtermHtmlTitle            string = "Cloudshell"
	defaultXtermKeepalivePingTimeout int    = 20
	defaultXtermMaxBufferSizeBytes   int    = 512
	defaultXtermUrlRoutePrefix       string = ""
	envarServerAddress               string = "SENZING_TOOLS_SERVER_ADDRESS"
	envarServerPort                  string = "SENZING_TOOLS_SERVER_PORT"
	envarXtermAllowedHostnames       string = "SENZING_TOOLS_XTERM_ALLOWED_HOSTNAMES"
	envarXtermArguments              string = "SENZING_TOOLS_XTERM_ARGUMENTS"
	envarXtermCommand                string = "SENZING_TOOLS_XTERM_COMMAND"
	envarXtermConnectionErrorLimit   string = "SENZING_TOOLS_XTERM_CONNECTION_ERROR_LIMIT"
	envarXtermHtmlTitle              string = "SENZING_TOOLS_XTERM_HTML_TITLE"
	envarXtermKeepalivePingTimeout   string = "SENZING_TOOLS_XTERM_KEEPALIVE_PING_TIMEOUT"
	envarXtermMaxBufferSizeBytes     string = "SENZING_TOOLS_XTERM_MAX_BUFFER_SIZE_BYTES"
	envarXtermUrlRoutePrefix         string = "SENZING_TOOLS_XTERM_URL_ROUTE_PREFIX"
	optionServerAddress              string = "server-addr"
	optionServerPort                 string = "server-port"
	optionXtermAllowedHostnames      string = "xterm-allowed-hostnames"
	optionXtermArguments             string = "xterm-arguments"
	optionXtermCommand               string = "xterm-command"
	optionXtermConnectionErrorLimit  string = "xterm-connection-error-limit"
	optionXtermHtmlTitle             string = "xterm-html-title"
	optionXtermKeepalivePingTimeout  string = "xterm-keepalive-ping-timeout"
	optionXtermMaxBufferSizeBytes    string = "xterm-max-buffer-size-bytes"
	optionXtermUrlRoutePrefix        string = "xterm-url-route-prefix"
	Short                            string = "view-xterm short description"
	Use                              string = "view-xterm"
	Long                             string = `
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
	RootCmd.Flags().Int(optionXtermConnectionErrorLimit, defaultXtermConnectionErrorLimit, fmt.Sprintf("Connection re-attempts before terminating [%s]", envarXtermConnectionErrorLimit))
	RootCmd.Flags().Int(optionXtermKeepalivePingTimeout, defaultXtermKeepalivePingTimeout, fmt.Sprintf("Maximum allowable seconds between a ping message and its response [%s]", envarXtermKeepalivePingTimeout))
	RootCmd.Flags().Int(optionXtermMaxBufferSizeBytes, defaultXtermMaxBufferSizeBytes, fmt.Sprintf("Maximum length of terminal input [%s]", envarXtermMaxBufferSizeBytes))
	RootCmd.Flags().Int(optionServerPort, defaultServerPort, fmt.Sprintf("Port the server listens on [%s]", envarServerPort))
	RootCmd.Flags().String(optionXtermCommand, defaultXtermCommand, fmt.Sprintf("Path of shell command [%s]", envarXtermCommand))
	RootCmd.Flags().String(optionXtermHtmlTitle, defaultXtermHtmlTitle, fmt.Sprintf("XTerm HTML page title [%s]", envarXtermHtmlTitle))
	RootCmd.Flags().String(optionServerAddress, defaultServerAddress, fmt.Sprintf("IP interface server listens on [%s]", envarServerAddress))
	RootCmd.Flags().String(optionXtermUrlRoutePrefix, defaultXtermUrlRoutePrefix, fmt.Sprintf("Route prefix [%s]", envarXtermUrlRoutePrefix))
	RootCmd.Flags().StringSlice(optionXtermAllowedHostnames, defaultAllowedHostnames, fmt.Sprintf("Comma-delimited list of hostnames permitted to connect to the websocket [%s]", envarXtermAllowedHostnames))
	RootCmd.Flags().StringSlice(optionXtermArguments, defaultArguments, fmt.Sprintf("Comma-delimited list of arguments passed to the terminal command prompt [%s]", envarXtermArguments))
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
		optionXtermConnectionErrorLimit: defaultXtermConnectionErrorLimit,
		optionXtermKeepalivePingTimeout: defaultXtermKeepalivePingTimeout,
		optionXtermMaxBufferSizeBytes:   defaultXtermMaxBufferSizeBytes,
		optionServerPort:                defaultServerPort,
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
		optionXtermCommand:        defaultXtermCommand,
		optionXtermHtmlTitle:      defaultXtermHtmlTitle,
		optionServerAddress:       defaultServerAddress,
		optionXtermUrlRoutePrefix: defaultXtermUrlRoutePrefix,
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
		optionXtermAllowedHostnames: defaultAllowedHostnames,
		optionXtermArguments:        defaultArguments,
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
		AllowedHostnames:     viper.GetStringSlice(optionXtermAllowedHostnames),
		Arguments:            viper.GetStringSlice(optionXtermArguments),
		Command:              viper.GetString(optionXtermCommand),
		ConnectionErrorLimit: viper.GetInt(optionXtermConnectionErrorLimit),
		HtmlTitle:            viper.GetString(optionXtermHtmlTitle),
		KeepalivePingTimeout: viper.GetInt(optionXtermKeepalivePingTimeout),
		MaxBufferSizeBytes:   viper.GetInt(optionXtermMaxBufferSizeBytes),
		ServerPort:           viper.GetInt(optionServerPort),
		ServerAddress:        viper.GetString(optionServerAddress),
		UrlRoutePrefix:       viper.GetString(optionXtermUrlRoutePrefix),
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
