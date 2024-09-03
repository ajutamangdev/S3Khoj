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
	"os/exec"
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

var regions = []string{
	"us-east-1", "us-east-2", "us-west-1", "us-west-2",
	"af-south-1", "ap-east-1", "ap-south-1", "ap-northeast-1",
	"ap-northeast-2", "ap-northeast-3", "ap-southeast-1",
	"ap-southeast-2", "ca-central-1", "cn-north-1", "cn-northwest-1",
	"eu-central-1", "eu-west-1", "eu-west-2", "eu-west-3",
	"eu-north-1", "eu-south-1", "me-south-1", "sa-east-1",
	"us-gov-east-1", "us-gov-west-1",
}

func createInsecureHTTPClient() *http.Client {
	//  skips certificate verification
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
		fmt.Printf("[-] Failed to list objects: %v\n", err)
		if strings.Contains(err.Error(), "AccessDenied") {
			fmt.Printf("[!] Access Denied: The bucket '%s' may be restricted despite being publicly accessible.\n", bucketName)
			fmt.Printf("[!] Verifying public access using AWS CLI: aws s3 ls %s --no-sign-request\n", bucketName)

			// Execute AWS CLI command
			cmd := exec.Command("aws", "s3", "ls", fmt.Sprintf("s3://%s", bucketName), "--no-sign-request")
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("[-] Error executing AWS CLI command: %v\n", err)
			} else {
				fmt.Printf("[+] AWS CLI output:\n%s\n", output)
			}
		}
	} else {
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
		for _, region := range regions {
			cfg, err := config.LoadDefaultConfig(context.TODO(),
				config.WithRegion(region),
				config.WithCredentialsProvider(aws.AnonymousCredentials{}),
			)
			if err != nil {
				log.Printf("failed to load config for region %s, %v", region, err)
				continue
			}

			svc := s3.NewFromConfig(cfg)

			_, err = svc.HeadBucket(context.TODO(), &s3.HeadBucketInput{
				Bucket: bucketName,
			})
			if err == nil {
				fmt.Printf("Bucket found in region %s\n", region)
				enumerateFiles(svc, *bucketName, commonFilesPatterns)
				break
			}
		}
	}
}
