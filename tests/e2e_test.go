package tests

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var logger, _ = zap.NewProduction()
var logType string

type LogResponse struct {
	Hits struct {
		Total int `json:"total"`
		Hits  []struct {
			Source struct {
				LogType  string `json:"type"`
				LogGroup string `json:"logGroup"`
				Message  string `json:"message"` // Add message field to check filter pattern
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func TestMain(m *testing.M) {
	// Get log type to search for
	flag.StringVar(&logType, "logType", "logzio_firehose", "log type to search for in logzio")
	flag.Parse()

	// Run the tests
	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestNeededDataGotToLogzio(t *testing.T) {
	logsApiToken := os.Getenv("LOGZIO_API_TOKEN")
	if logsApiToken == "" {
		t.Fatalf("LOGZIO_LOGS_API_KEY environment variable not set")
	}
	logzioLogs, err := fetchLogs(logsApiToken)
	if err != nil {
		t.Fatalf("Failed to fetch logs: %v", err)
	}

	if logzioLogs.Hits.Total == 0 {
		t.Errorf("No logs found")
	}

	possibleLogGroups := []string{"API-Gateway-Execution-Logs_", "/aws/apigateway/welcome"}
	for _, hit := range logzioLogs.Hits.Hits {
		assert.True(t, containsAny(hit.Source.LogGroup, possibleLogGroups))
	}

}

func TestFilterPattern(t *testing.T) {
	logsApiToken := os.Getenv("LOGZIO_API_TOKEN")
	if logsApiToken == "" {
		t.Fatalf("LOGZIO_LOGS_API_KEY environment variable not set")
	}

	logzioLogs, err := fetchLogs(logsApiToken)
	if err != nil {
		t.Fatalf("Failed to fetch logs: %v", err)
	}

	assert.Greater(t, logzioLogs.Hits.Total, 0, "No logs found after applying filter pattern")

	for _, hit := range logzioLogs.Hits.Hits {
		assert.Contains(t, hit.Source.Message, "GET",
			fmt.Sprintf("Log doesn't contain 'GET' which should be required by filter pattern: %s", hit.Source.Message))
	}
}

func fetchLogs(logsApiToken string) (*LogResponse, error) {
	url := "https://api.logz.io/v1/search"
	client := &http.Client{}
	query := fmt.Sprintf(`{
		"query": {
			"query_string": {
				"query": "type:%s"
			}
		}
	}`, logType)

	logger.Info("sending api request", zap.String("url", url), zap.String("query", query))
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(query))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-TOKEN", logsApiToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var logResponse LogResponse
	err = json.Unmarshal(body, &logResponse)
	if err != nil {
		return nil, err
	}

	return &logResponse, nil
}

func containsAny(s string, subStrings []string) bool {
	for _, subStr := range subStrings {
		if strings.Contains(s, subStr) {
			return true
		}
	}
	return false
}
