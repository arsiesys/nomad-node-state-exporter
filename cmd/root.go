/*
Copyright © 2022 Loïc Yavercovski

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	exporter "github.com/arsiesys/nomad-node-state-exporter/pkg/exporter"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nomad-node-state-exporter",
	Short: "Generate prometheus metrics for nomad nodes states",
	Long:  `Generate prometheus metrics for nomad nodes states`,
	Run: func(cmd *cobra.Command, args []string) {
		exporter.Entrypoint()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringP("address", "a", "https://my-nomad-server:4646", "address of the nomad server api")
	if err := viper.BindPFlag("address", rootCmd.Flags().Lookup("address")); err != nil {
		log.Fatal(err)
	}
	rootCmd.Flags().Int("port", 9827, "port to listen on")
	if err := viper.BindPFlag("port", rootCmd.Flags().Lookup("port")); err != nil {
		log.Fatal(err)
	}
	rootCmd.Flags().Duration("fetch-interval", 30*time.Second, "fetch-interval in seconds")
	if err := viper.BindPFlag("fetch-interval", rootCmd.Flags().Lookup("fetch-interval")); err != nil {
		log.Fatal(err)
	}
	rootCmd.Flags().Bool("disable-authentication", false, "disable authentication")
	if err := viper.BindPFlag("disable-authentication", rootCmd.Flags().Lookup("disable-authentication")); err != nil {
		log.Fatal(err)
	}
	rootCmd.Flags().StringP("cert", "", "/nomad-pki/cli.pem", "Certificate used for TLS auth")
	if err := viper.BindPFlag("cert", rootCmd.Flags().Lookup("cert")); err != nil {
		log.Fatal(err)
	}
	rootCmd.Flags().StringP("key", "", "/nomad-pki/cli-key.pem", "Certificate KEY used for TLS auth")
	if err := viper.BindPFlag("key", rootCmd.Flags().Lookup("key")); err != nil {
		log.Fatal(err)
	}
	rootCmd.Flags().StringP("ca", "", "/nomad-pki/nomad-ca.pem", "Trusting CA certificate for TLS auth")
	if err := viper.BindPFlag("ca", rootCmd.Flags().Lookup("ca")); err != nil {
		log.Fatal(err)
	}
	rootCmd.Flags().StringP("filter", "f", "", "Nomad format expression filter for allocations endpoint. example: Name contains \"jenkins\"")
	if err := viper.BindPFlag("filter", rootCmd.Flags().Lookup("filter")); err != nil {
		log.Fatal(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		// Search config in home directory with name ".nomad-node-state-exporter" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".nomad-node-state-exporter")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	log.Printf("Using nomad api: %s", viper.GetString("address"))
	if viper.GetBool("disable-authentication") {
		log.Printf("Authentication disabled")
	} else {
		log.Printf("Using TLS cert: %s", viper.GetString("cert"))
		log.Printf("Using TLS key: %s", viper.GetString("key"))
		log.Printf("Using TLS ca: %s", viper.GetString("ca"))
	}
	log.Printf("Listening on port: %d", viper.GetInt("port"))
	log.Printf("Parsing interval: %fs", viper.GetDuration("fetch-interval").Seconds())
	log.Printf("Using allocators filter: %s", viper.GetString("filter"))
}
