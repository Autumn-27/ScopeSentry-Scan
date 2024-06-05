package core

import (
	"regexp"
	"strings"
)

const SUBRE = `(?i)(([a-zA-Z0-9]{1}|[_a-zA-Z0-9]{1}[_a-zA-Z0-9-]{0,61}[a-zA-Z0-9]{1})[.]{1})+`

var AWSS3 = regexp.MustCompile(`(?i)[a-z0-9.-]+\.s3\.amazonaws\.com|[a-z0-9.-]+\.s3-[a-z0-9-]\.amazonaws\.com|[a-z0-9.-]+\.s3-website[.-](eu|ap|us|ca|sa|cn)|//s3\.amazonaws\.com/[a-z0-9._-]+|//s3-[a-z0-9-]+\.amazonaws\.com/[a-z0-9._-]+`)

// SubdomainRegex returns a Regexp object initialized to match
// subdomain names that end with the domain provided by the parameter.
func subdomainRegex(domain string) *regexp.Regexp {
	// Change all the periods into literal periods for the regex
	d := strings.Replace(domain, ".", "[.]", -1)
	return regexp.MustCompile(SUBRE + d)
}

func GetSubdomains(source, domain string) []string {
	chunkSize := 5120
	overlapSize := 100
	var subs []string
	for start := 0; start < len(source); start += chunkSize {
		end := start + chunkSize
		if end > len(source) {
			end = len(source)
		}

		chunkEnd := end
		if end+overlapSize < len(source) {
			chunkEnd = end + overlapSize
		}
		re := subdomainRegex(domain)
		for _, match := range re.FindAllStringSubmatch(source[start:chunkEnd], -1) {
			subs = append(subs, CleanSubdomain(match[0]))
		}
	}
	return subs
}

func GetAWSS3(source string) []string {
	chunkSize := 5120
	overlapSize := 100
	var aws []string
	for start := 0; start < len(source); start += chunkSize {
		end := start + chunkSize
		if end > len(source) {
			end = len(source)
		}
		chunkEnd := end
		if end+overlapSize < len(source) {
			chunkEnd = end + overlapSize
		}
		for _, match := range AWSS3.FindAllStringSubmatch(source[start:chunkEnd], -1) {
			aws = append(aws, match[0])
		}
	}
	return aws
}
