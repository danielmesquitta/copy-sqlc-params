/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/danielmesquitta/copy-sqlc-params/internal/domain/usecase"
	"github.com/spf13/cobra"
)

var inDir, outDir, packageName string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "copy-sqlc-params",
	Short: "Copy SQLC generated params to a new file",
	Long:  `Copy SQLC generated params to a new file`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		outFile, err := usecase.CopySQLCParams(inDir, outDir, packageName)
		if err != nil {
			fmt.Println("copy-sqlc-params: ", err)
			os.Exit(1)
		}

		fmt.Printf("copy-sqlc-params: wrote %s\n", outFile)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().
		StringVarP(&inDir, "input", "i", "", "Path to the input dir with sqlc generated files")
	rootCmd.Flags().
		StringVarP(&outDir, "output", "o", "", "Path to the output dir for the generated file")
	rootCmd.Flags().
		StringVarP(&packageName, "package", "p", "", "Package name for the generated file")

	if err := rootCmd.MarkFlagRequired("input"); err != nil {
		panic(err)
	}
	if err := rootCmd.MarkFlagRequired("output"); err != nil {
		panic(err)
	}
}
