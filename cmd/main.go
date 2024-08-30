package cmd

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Pattern struct {
	Match   string
	IsRegex bool
}

func checkBucketPublic(bucketName string) bool {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s.s3.amazonaws.com/", bucketName), nil)
	if err != nil {
		fmt.Printf("[-] Error creating request: %v\n", err)
		return false
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[-] Error making request: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("[+] Bucket '%s' is publicly accessible.\n", bucketName)
		return true
	} else {
		fmt.Printf("[-] Bucket '%s' is not publicly accessible. Status code: %d\n", bucketName, resp.StatusCode)
		return false
	}
}

func loadCommonFiles(filename string) ([]Pattern, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("[-] Error opening common files list: %v", err)
	}
	defer file.Close()

	var patterns []Pattern
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pattern := scanner.Text()
		isRegex := strings.HasPrefix(pattern, "regex:")

		if isRegex {
			pattern = strings.TrimPrefix(pattern, "regex:")
		}

		patterns = append(patterns, Pattern{
			Match:   pattern,
			IsRegex: isRegex,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("[-] Error reading common files list: %v", err)
	}

	return patterns, nil
}

func enumerateFiles(svc *s3.Client, bucketName string, patterns []Pattern) {
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	resp, err := svc.ListObjectsV2(context.TODO(), params)
	if err != nil {
		log.Fatalf("Failed to list objects: %v", err)
	}

	for _, item := range resp.Contents {
		for _, pattern := range patterns {
			if pattern.IsRegex {
				matched, err := regexp.MatchString(pattern.Match, *item.Key)
				if err != nil {
					fmt.Printf("[-] Error matching regex: %v\n", err)
					continue
				}
				if matched {
					fmt.Printf("[+] Found matching file '%s' in bucket '%s'\n", *item.Key, bucketName)
				}
			} else if strings.Contains(*item.Key, pattern.Match) {
				fmt.Printf("[+] Found matching file '%s' in bucket '%s'\n", *item.Key, bucketName)
			}
		}
	}
}

func runMain() {
	bucketName := flag.String("b", "", "The name of the S3 bucket")
	sourceFile := flag.String("s", "", "External directory list file")
	flag.Parse()

	if *bucketName == "" {
		fmt.Println("Error: You must provide a bucket name using the -b flag")
		return
	}

	var commonFilesPatterns []Pattern
	var err error

	if *sourceFile != "" {
		commonFilesPatterns, err = loadCommonFiles(*sourceFile)
	} else {
		commonFilesPatterns, err = loadCommonFiles("config/common-files.txt")
	}

	if err != nil {
		fmt.Println("[-] Error loading common files:", err)
		return
	}

	if checkBucketPublic(*bucketName) {
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion("us-east-1"),
			config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		)
		if err != nil {
			log.Fatalf("failed to load config, %v", err)
		}
		svc := s3.NewFromConfig(cfg)

		enumerateFiles(svc, *bucketName, commonFilesPatterns)
	}
}
