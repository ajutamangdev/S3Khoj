package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func printBanner() {
	fmt.Println(`
  _________________  ____  __.__               __ 
 /   _____/\_____  \|    |/ _|  |__   ____    |__|
 \_____  \   _(__  <|      < |  |  \ /  _ \   |  |
 /        \ /       \    |  \|   Y  (  <_> )  |  |
/_______  //______  /____|__ \___|  /\____/\__|  |
        \/        \/        \/    \/      \______|

S3 inspector tool that help pentesters to extract juicy information from the public accessible S3 buckets.
	`)
}

var (
	bucketName       string
	externalFileList string
)

var rootCmd = &cobra.Command{
	Use:   "S3Khoj",
	Short: "S3Khoj is a inspector tool that help pentesters to extract juicy information from the public accessible S3 buckets.",
	Long:  "S3Khoj is a inspector tool that help pentesters to extract juicy information from the public accessible S3 buckets.",
	Run: func(cmd *cobra.Command, args []string) {
		printBanner()
		runMain()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&bucketName, "bucket", "b", "", "Name of the s3 bucket to check")
	rootCmd.PersistentFlags().StringVarP(&externalFileList, "source", "s", "", "External directory list file")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing S3Khoj '%s'\n", err)
		os.Exit(1)
	}
}
