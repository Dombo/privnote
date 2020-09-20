package cmd

import (
	"fmt"
	"github.com/dombo/privnote/lib"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

var (
	// Used for flags.
	cfgFile     string
	doNotPrompt bool
	password	bool
	expires     = map[string]map[string]string{
		"0":   {"usage": "expire after 1st read, or 30 days in unread state", "hours": "0"},
		"1h":  {"usage": "1 hour", "hours": "1"},
		"24h": {"usage": "24 hours", "hours": "24"}, "1d": {"usage": "1 day", "hours": "24"},
		"7d": {"usage": "7 days", "hours": "168"}, "1w": {"usage": "1 week", "hours": "168"},
		"30d": {"usage": "30 days", "hours": "720"}, "1m": {"usage": "1 month", "hours": "720"},
	}

	usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}
{{if .HasAvailableLocalFlags}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}

Use "{{.CommandPath}} completion --help" for more information about enabling shell completions.
`

	rootCmd = &cobra.Command{
		Use: "privnote",
		Short: "A utility for sharing one-time-read secrets via privnote.com",
		Long: "Share secrets with third parties securely over questionable communication channels via privnote.com",
		PreRunE: func(cmd *cobra.Command, args []string) error { // Validate flags and arguments
			var errors []string

			if cmd.Flags().Lookup("expires").Changed == true {
				passedValue := cmd.Flags().Lookup("expires").Value.String()
				if expires[passedValue]["usage"] == "" {
					errors = append(errors,
						fmt.Errorf("\n\tinvalid value passed to expires: %v\n", passedValue).Error(),
					)
				}
			}
			if cmd.Flags().Lookup("file").Changed == true {
				passedFile := cmd.Flags().Lookup("file").Value.String()
				if _, err := os.Stat(passedFile); err != nil {
					if os.IsNotExist(err) {
						errors = append(errors,
							fmt.Errorf("\n\tpath passed to file does not exist: %v\n", passedFile).Error(),
						)
					}
					if os.IsPermission(err) {
						errors = append(errors,
							fmt.Errorf("\n\tpath passed to file cannot be read: %v\n", passedFile).Error(),
						)
					}
				}
			} else {
				piped, err := os.Stdin.Stat()
				if err != nil {
					log.Panicf("failed to open pipe %v", err)
				}
				if piped.Mode() & os.ModeNamedPipe == 0 {
					errors = append(errors,
						fmt.Errorf("\n\tyou must specify something to encrypt via pipe or file flag\n").Error(),
					)
				}
			}

			if len(errors) == 0 {
				return nil
			}

			return fmt.Errorf("invalid arguments specified, unrecoverable: %v\n", errors)
		},
		Run: func(cmd *cobra.Command, args []string) {
			lib.CreateNote(cmd, expires[viper.GetString("expires")]["hours"])
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			var completions []string
			// We manually build the completion suggestions array off the root command
			cmd.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
				if strings.HasPrefix(flag.Name, toComplete) {
					completions = append(completions, []string{fmt.Sprintf("--%s\t%s", flag.Name, flag.Usage)}...)
					if flag.Shorthand != "" {
						completions = append(completions, []string{fmt.Sprintf("-%s\t%s", flag.Shorthand, flag.Usage)}...)
					}
				}
			})

			return completions, cobra.ShellCompDirectiveNoFileComp // Prevent file completion for better DX
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVarP(&doNotPrompt, "do-not-prompt", "", false, "do not prompt the receiver before they open the note that it is one time read")
	rootCmd.PersistentFlags().StringP("expires", "e", "0", "note destroyed automatically after specified period")
	rootCmd.PersistentFlags().StringP("file", "f", "", "file to encrypt and store in the privnote, piped input takes priority")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config-file", "c", "", "config file to override defaults, otherwise allows a .privnote stored in home dir")
	rootCmd.PersistentFlags().StringP("notify-email", "", "", "email to receive notification on note open")
	rootCmd.PersistentFlags().StringP("notify-reference", "", "", "reference included in notification on note open")
	rootCmd.PersistentFlags().BoolVarP(&password, "password", "p", false, "specify a password that must be entered before someone can your note")

	viper.BindPFlag("do-not-prompt", rootCmd.PersistentFlags().Lookup("do-not-prompt"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("expires", rootCmd.PersistentFlags().Lookup("expires"))
	viper.BindPFlag("notify-email", rootCmd.PersistentFlags().Lookup("notify-email"))

 	rootCmd.RegisterFlagCompletionFunc("expires", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		for k, v := range expires {
			completions = append(completions, fmt.Sprintf("%s\t%s", k, v["usage"]))
		}
		return completions, cobra.ShellCompDirectiveDefault
	})
	rootCmd.SetUsageTemplate(usageTemplate)
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			er(err)
		}

		// Search config in home directory with name ".privnote" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".privnote")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}