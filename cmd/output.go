package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Result struct {
	BucketName    string   `json:"bucket_name"`
	Region        string   `json:"region"`
	IsPublic      bool     `json:"is_public"`
	Files         []string `json:"files"`
	MatchingFiles []string `json:"matching_files"`
}

func outputJSON(result Result) {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}
	fmt.Println(string(jsonData))
}

func exportJSON(result Result) error {
	jsonFileName := result.BucketName + ".json"
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	err = os.WriteFile(jsonFileName, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing JSON file: %v", err)
	}

	fmt.Printf("JSON configuration exported to %s\n", jsonFileName)
	return nil
}

func outputCSV(result Result) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	writeCSV(writer, result)
}

func exportCSV(result Result) error {
	csvFileName := result.BucketName + ".csv"
	csvFile, err := os.Create(csvFileName)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	writeCSV(writer, result)

	if err := writer.Error(); err != nil {
		return fmt.Errorf("error writing CSV: %v", err)
	}

	fmt.Printf("CSV configuration exported to %s\n", csvFileName)
	return nil
}

func writeCSV(writer *csv.Writer, result Result) {
	// Write header
	writer.Write([]string{"Bucket", "Region", "Is Public", "File"})

	// Write data
	if len(result.Files) > 0 {
		for _, file := range result.Files {
			writer.Write([]string{result.BucketName, result.Region, fmt.Sprintf("%t", result.IsPublic), file})
		}
	} else {
		writer.Write([]string{result.BucketName, result.Region, fmt.Sprintf("%t", result.IsPublic), ""})
	}
}

func ExportJSON(result Result) error {
	jsonFileName := result.BucketName + ".json"
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	err = os.WriteFile(jsonFileName, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing JSON file: %v", err)
	}

	fmt.Printf("JSON configuration exported to %s\n", jsonFileName)
	return nil
}

func ExportCSV(result Result) error {
	csvFileName := result.BucketName + ".csv"
	csvFile, err := os.Create(csvFileName)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	writeCSV(writer, result)

	if err := writer.Error(); err != nil {
		return fmt.Errorf("error writing CSV: %v", err)
	}

	fmt.Printf("CSV configuration exported to %s\n", csvFileName)
	return nil
}

func ExportHTML(result Result) error {
	htmlFileName := result.BucketName + ".html"
	htmlFile, err := os.Create(htmlFileName)
	if err != nil {
		return fmt.Errorf("error creating HTML file: %v", err)
	}
	defer htmlFile.Close()

	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>S3 Bucket Enumeration Results</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; padding: 20px; }
        h1 { color: #333; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        a { color: #0066cc; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>S3 Bucket Enumeration Results</h1>
    <p><strong>Bucket Name:</strong> {{.BucketName}}</p>
    <p><strong>Region:</strong> {{.Region}}</p>
    <p><strong>Publicly Accessible:</strong> {{.IsPublic}}</p>
    {{if .IsPublic}}
        <h2>Files Found:</h2>
        {{if .Files}}
            <table>
                <tr>
                    <th>File Name</th>
                </tr>
                {{range .Files}}
                <tr>
                    <td><a href="https://{{$.BucketName}}.s3.{{$.Region}}.amazonaws.com/{{.}}" target="_blank">{{.}}</a></td>
                </tr>
                {{end}}
            </table>
        {{else}}
            <p>No files found.</p>
        {{end}}
    {{end}}
</body>
</html>
`

	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("error parsing HTML template: %v", err)
	}

	err = t.Execute(htmlFile, result)
	if err != nil {
		return fmt.Errorf("error executing HTML template: %v", err)
	}

	fmt.Printf("HTML configuration exported to %s\n", htmlFileName)
	return nil
}

func DownloadPublicFiles(result Result, s3Client *s3.Client) error {
	downloadDir := fmt.Sprintf("%s_downloads", result.BucketName)
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return fmt.Errorf("error creating download directory: %v", err)
	}

	fmt.Printf("Matching files: %v\n", result.MatchingFiles)
	fmt.Printf("All files: %v\n", result.Files)

	// Create a map of matching files for quick lookup
	matchingFiles := make(map[string]bool)
	for _, file := range result.MatchingFiles {
		matchingFiles[file] = true
	}

	downloadCount := 0
	for _, file := range result.Files {
		// Only download if it's a matching file
		if !matchingFiles[file] {
			fmt.Printf("Skipping non-matching file: %s\n", file)
			continue
		}

		// Use s3Client to download the file
		input := &s3.GetObjectInput{
			Bucket: aws.String(result.BucketName),
			Key:    aws.String(file),
		}
		output, err := s3Client.GetObject(context.TODO(), input)
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", file, err)
			continue
		}

		localPath := filepath.Join(downloadDir, file)
		dir := filepath.Dir(localPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			continue
		}

		localFile, err := os.Create(localPath)
		if err != nil {
			fmt.Printf("Error creating local file %s: %v\n", localPath, err)
			continue
		}
		defer localFile.Close()

		_, err = io.Copy(localFile, output.Body)
		output.Body.Close()
		if err != nil {
			fmt.Printf("Error writing to local file %s: %v\n", localPath, err)
			continue
		}
		fmt.Printf("Downloaded: %s\n", localPath)
		downloadCount++
	}

	fmt.Printf("Total files downloaded: %d\n", downloadCount)

	return nil
}
