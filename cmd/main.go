package cmd

import (
	"bufio"
	"context"
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

func listObjects(bucketName string) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		log.Fatalf("failed to load config, %v", err)
	}

	svc := s3.NewFromConfig(cfg)

	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	resp, err := svc.ListObjectsV2(context.TODO(), params)
	if err != nil {
		log.Fatalf("Failed to list objects: %v", err)
	}

	for _, item := range resp.Contents {
		fmt.Println("Object:", *item.Key)
	}
}

func enumerateCommonFiles(svc *s3.Client, bucketName string, commonFiles []string) {
	for _, file := range commonFiles {
		params := &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
			Prefix: aws.String(file),
		}

		resp, err := svc.ListObjectsV2(context.TODO(), params)
		if err != nil {
			fmt.Printf("[-] Error listing objects for '%s' in bucket '%s': %v\n", file, bucketName, err)
			continue
		}

		for _, item := range resp.Contents {
			if strings.Contains(*item.Key, file) {
				fmt.Printf("[+] Found '%s' in bucket '%s'\n", *item.Key, bucketName)
			}
		}
	}
}

func runMain() {
	if bucketName == "" {
		fmt.Println("Error: You must provide a bucket name")
		return
	}

	commonFiles, err := readFiles("common-files.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	if checkBucketPublic(bucketName) {

		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion("us-east-1"),
			config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		)
		if err != nil {
			log.Fatalf("failed to load config, %v", err)
		}
		svc := s3.NewFromConfig(cfg)

		//listObjects(bucketName)
		enumerateCommonFiles(svc, bucketName, commonFiles)
	}
}

func readFiles(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("[-] Error opening sensitive files list: %v", err)
	}
	defer file.Close()

	var commonFiles []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		commonFiles = append(commonFiles, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("[-] Error reading common files list: %v", err)
	}

	return commonFiles, nil
}
