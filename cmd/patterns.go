package cmd

import (
	"fmt"
	"regexp"
)

type Pattern struct {
	Match         string
	IsRegex       bool
	CompiledRegex *regexp.Regexp
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

var DefaultPatterns = []Pattern{
	{Match: "^.*\\.git", IsRegex: false},
	{Match: "^.*\\.tfstate", IsRegex: true},
	{Match: "^.*\\.env$", IsRegex: true},
	{Match: "^.*\\.xml$", IsRegex: true},
	{Match: "^.*\\.log", IsRegex: true},
	{Match: "^.*\\.swp", IsRegex: true},
	{Match: "^.*\\.sql", IsRegex: true},
	{Match: "^.*\\.zip", IsRegex: true},
	{Match: "^.*\\.pem", IsRegex: true},
	{Match: "^.*\\.conf", IsRegex: true},
	{Match: "^.*\\.ini", IsRegex: true},
	{Match: "^.*\\.bak", IsRegex: true},
	{Match: "^.*\\.xls", IsRegex: true},
	{Match: "^.*\\.txt", IsRegex: true},
	{Match: "^.*\\.csv", IsRegex: true},
	{Match: "^.*\\.db", IsRegex: true},
	{Match: "^.*\\.sqlite", IsRegex: true},
	{Match: "^.*\\tar.gz", IsRegex: true},
	{Match: "^.*\\.properties", IsRegex: true},
	{Match: "^.*\\.dist", IsRegex: true},
	{Match: "^.*\\.cfg", IsRegex: true},
	{Match: "^.*\\.mdb", IsRegex: true},
	{Match: "^.*\\.key", IsRegex: true},
	{Match: "^.*\\.json", IsRegex: true},
	{Match: "^.*\\.htaccess", IsRegex: true},
	{Match: "^.*\\.htpasswd", IsRegex: true},
	{Match: "^.*\\.lock", IsRegex: true},
	{Match: "^.*\\.settings", IsRegex: true},
	{Match: "^.*\\.project", IsRegex: true},
	{Match: "^.*\\.swf", IsRegex: true},
	{Match: "^.*\\.inc", IsRegex: true},
	{Match: "^.*\\tmp", IsRegex: true},
	{Match: "^.*\\temp", IsRegex: true},
	{Match: "^.*\\.html", IsRegex: true},
}

func CompilePatterns(patterns []Pattern) []Pattern {
	for i, p := range patterns {
		if p.IsRegex {
			compiled, err := regexp.Compile(p.Match)
			if err != nil {
				fmt.Printf("[-] Error compiling regex '%s': %v\n", p.Match, err)
				continue
			}
			patterns[i].CompiledRegex = compiled
		}
	}
	return patterns
}
