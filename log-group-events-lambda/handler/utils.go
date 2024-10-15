package handler

import (
	"strings"
)

func getServicesMap() map[string]string {
	return map[string]string{
		"apigateway":       "/aws/apigateway/",
		"rds":              "/aws/rds/",
		"cloudhsm":         "/aws/cloudhsm/",
		"cloudtrail":       "aws-cloudtrail-logs-",
		"codebuild":        "/aws/codebuild/",
		"connect":          "/aws/connect/",
		"elasticbeanstalk": "/aws/elasticbeanstalk/",
		"ecs":              "/aws/ecs/",
		"eks":              "/aws/eks/",
		"aws-glue":         "/aws/aws-glue/",
		"aws-iot":          "AWSIotLogsV2",
		"lambda":           "/aws/lambda/",
		"macie":            "/aws/macie/",
		"amazon-mq":        "/aws/amazonmq/broker/",
		"batch":            "/aws/batch/",
	}
}

func convertStrToArr(s string) []string {
	if s == emptyString {
		return nil
	}

	s = strings.ReplaceAll(s, " ", "")
	return strings.Split(s, valuesSeparator)
}

// findDifferences finds elements in 'new' that are not in 'old', and vice versa.
func findDifferences(old, new []string) (toAdd, toRemove []string) {
	oldSet := make(map[string]struct{})
	newSet := make(map[string]struct{})

	// Populate 'oldSet' with elements from the 'old' slice.
	for _, item := range old {
		oldSet[item] = struct{}{}
	}

	for _, item := range new {
		newSet[item] = struct{}{}
	}

	// Find elements in 'new' that are not in 'old' and add them to 'toAdd'.
	for item := range newSet {
		_, exists := oldSet[item] // Check if 'item' exists in 'oldSet'
		if !exists {
			toAdd = append(toAdd, item)
		}
	}

	for item := range oldSet {
		_, exists := newSet[item]
		if !exists {
			toRemove = append(toRemove, item)
		}
	}

	return toAdd, toRemove
}
