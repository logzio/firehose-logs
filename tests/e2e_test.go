package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"testing"
)

var logger, _ = zap.NewProduction()

type LogResponse struct {
	Hits struct {
		Total int `json:"total"`
		Hits  []struct {
			Source struct {
				LogType string `json:"type"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
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

	/* To add extra checks here */
}

func fetchLogs(logsApiToken string) (*LogResponse, error) {
	url := "https://api.logz.io/v1/search"
	client := &http.Client{}
	query := `{
		"query": {
			"query_string": {
				"query": "type:logzio_firehose"
			}
		}
	}`

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
