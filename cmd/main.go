package cmd

import (
	"bufio"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func createInsecureHTTPClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr, Timeout: 10 * time.Second}
}

func checkBucketPublic(bucketName string) bool {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s.s3.amazonaws.com/", bucketName), nil)
	if err != nil {
		fmt.Printf("[-] Error creating request: %v\n", err)
		return false
	}

	client := createInsecureHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[-] Error making request: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("[+] Bucket '%s' is publicly accessible.\n", bucketName)
		return true
	} else if resp.StatusCode == http.StatusForbidden {
		fmt.Printf("[+] Bucket '%s' is publicly accessible but access is restricted. Status code: %d\n", bucketName, resp.StatusCode)
		return true
	} else {
		fmt.Printf("[-] Bucket '%s' is not publicly accessible. Status code: %d\n", bucketName, resp.StatusCode)
		return false
	}
}

func loadCommonFiles(filename string) ([]Pattern, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening common files list: %v", err)
	}
	defer file.Close()

	var patterns []Pattern
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		isRegex := strings.HasPrefix(line, "regex:")
		if isRegex {
			line = strings.TrimPrefix(line, "regex:")
		}
		patterns = append(patterns, Pattern{Match: line, IsRegex: isRegex})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading common files list: %v", err)
	}

	return CompilePatterns(patterns), nil
}

func enumerateFiles(svc *s3.Client, bucketName string, patterns []Pattern) Result {
	result := Result{
		BucketName:    bucketName,
		MatchingFiles: []string{},
		Files:         []string{},
	}

	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	var foundFiles []string

	resp, err := svc.ListObjectsV2(context.TODO(), params)
	if err != nil {
		fmt.Printf("[-] Failed to list objects: %v\n", err)
		return result
	}

	for _, item := range resp.Contents {
		result.Files = append(result.Files, *item.Key)
		foundFiles = append(foundFiles, *item.Key)
		for _, pattern := range patterns {
			if pattern.IsRegex {
				if pattern.CompiledRegex != nil && pattern.CompiledRegex.MatchString(*item.Key) {
					fmt.Printf("[+] Found matching file '%s' in bucket '%s'\n", *item.Key, bucketName)
					result.MatchingFiles = append(result.MatchingFiles, *item.Key)
				}
			} else if strings.Contains(*item.Key, pattern.Match) {
				fmt.Printf("[+] Found matching file '%s' in bucket '%s'\n", *item.Key, bucketName)
				result.MatchingFiles = append(result.MatchingFiles, *item.Key)
			}
		}
	}

	return result
}

func runMain() {
	bucketName := flag.String("b", "", "The name of the S3 bucket")
	sourceFile := flag.String("w", "", "Custom Wordlist configuration file")
	outputFormat := flag.String("o", "text", "Output format: text, json, csv, or html")
	downloadFiles := flag.Bool("d", false, "Download all public files")

	flag.Parse()

	if *bucketName == "" {
		fmt.Println("Error: You must provide a bucket name using the -b flag")
		return
	}

	var commonFilesPatterns []Pattern
	var err error

	if *sourceFile != "" {
		// Load patterns from custom file if provided
		commonFilesPatterns, err = loadCommonFiles(*sourceFile)
		if err != nil {
			fmt.Println("[-] Error loading custom patterns:", err)
			return
		}
	} else {
		// Use DefaultPatterns if no custom file is provided
		commonFilesPatterns = CompilePatterns(DefaultPatterns)
	}

	result := Result{
		BucketName: *bucketName,
		IsPublic:   false,
		Files:      []string{},
	}

	result.IsPublic = checkBucketPublic(*bucketName)

	var svc *s3.Client // Declare svc here

	if result.IsPublic {
		for _, region := range regions {
			cfg, err := config.LoadDefaultConfig(context.TODO(),
				config.WithRegion(region),
				config.WithCredentialsProvider(aws.AnonymousCredentials{}),
			)
			if err != nil {
				log.Printf("failed to load config for region %s, %v", region, err)
				continue
			}

			svc = s3.NewFromConfig(cfg) // Assign svc here

			_, err = svc.HeadBucket(context.TODO(), &s3.HeadBucketInput{
				Bucket: bucketName,
			})
			if err == nil {
				fmt.Printf("Bucket found in region %s\n", region)
				result.Region = region
				enumResult := enumerateFiles(svc, *bucketName, commonFilesPatterns)
				result.MatchingFiles = enumResult.MatchingFiles
				result.Files = enumResult.Files
				break
			}
		}
	}

	// Export results based on the specified format
	switch *outputFormat {
	case "json":
		if err := ExportJSON(result); err != nil {
			fmt.Printf("Error exporting JSON: %v\n", err)
		}
	case "csv":
		if err := ExportCSV(result); err != nil {
			fmt.Printf("Error exporting CSV: %v\n", err)
		}
	case "html":
		if err := ExportHTML(result); err != nil {
			fmt.Printf("Error exporting HTML: %v\n", err)
		}
	default:

	}

	if *downloadFiles && svc != nil {
		if err := DownloadPublicFiles(result, svc); err != nil {
			fmt.Printf("Error downloading files: %v\n", err)
		}
	}
}
