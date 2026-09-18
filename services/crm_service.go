package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"new-pos-api/models"
	"os"
	"time"
)

func SyncHistory(input models.ScoreHistory) (int, []byte, error) {
	endpoint := os.Getenv("CRM_ENDPOINT")
	if endpoint == "" {
		return 0, nil, errors.New("missing CRM_ENDPOINT in environment")
	}

	url := endpoint + "/api/score/history/sync"
	method := "POST"

	// Marshal input to JSON
	jsonBody, err := json.Marshal(input)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	payload := bytes.NewReader(jsonBody)

	// Set timeout for HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	auth := os.Getenv("JOYLIDAY_AUTH")
	if auth == "" {
		return 0, nil, errors.New("missing JOYLIDAY_AUTH in environment")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+auth)

	// Execute request
	res, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	// Read response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle HTTP errors
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.StatusCode, body, fmt.Errorf("request failed with status %s", res.Status)
	}

	return res.StatusCode, body, nil
}

func GetCustomerByMobileNo(tel string) (int, *models.CustomerCrm, error) {
	endpoint := os.Getenv("CRM_ENDPOINT")
	if endpoint == "" {
		return 0, nil, errors.New("missing CRM_ENDPOINT in environment")
	}
	url := endpoint + "/api/customer/check/mobile-no/" + tel
	method := "GET"

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	auth := os.Getenv("JOYLIDAY_AUTH")
	if auth == "" {
		return 0, nil, errors.New("missing JOYLIDAY_AUTH in environment")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+auth)

	res, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.StatusCode, nil, fmt.Errorf("request failed with status %s", res.Status)
	}

	var resp models.CustomerCrmResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return res.StatusCode, nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return res.StatusCode, &resp.Data, nil
}

func fetchBranchByCodeFromCRM(code string) (int, *models.BranchCrm, error) {
	endpoint := os.Getenv("CRM_ENDPOINT")
	if endpoint == "" {
		return 0, nil, errors.New("missing CRM_ENDPOINT in environment")
	}
	url := endpoint + "/api/branch/by-code/" + code
	method := "GET"

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	auth := os.Getenv("JOYLIDAY_AUTH")
	if auth == "" {
		return 0, nil, errors.New("missing JOYLIDAY_AUTH in environment")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+auth)

	res, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.StatusCode, nil, fmt.Errorf("request failed with status %s", res.Status)
	}

	var resp models.BranchCrmResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return res.StatusCode, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return res.StatusCode, &resp.Data, nil
}

func fetchScoreTypesFromCRM() (int, []models.ScoreType, error) {
	endpoint := os.Getenv("CRM_ENDPOINT")
	if endpoint == "" {
		return 0, nil, errors.New("missing CRM_ENDPOINT in environment")
	}
	url := endpoint + "/api/score/type"
	method := "GET"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	auth := os.Getenv("JOYLIDAY_AUTH")
	if auth == "" {
		return 0, nil, errors.New("missing JOYLIDAY_AUTH in environment")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+auth)

	res, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.StatusCode, nil, fmt.Errorf("request failed with status %s", res.Status)
	}

	var resp models.ScoreTypeCrmResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return res.StatusCode, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return res.StatusCode, resp.Data, nil
}

func DeleteHistoryByRefTransaction(ref_transaction string, ref_transaction_id string) (int, *models.RowCount, error) {
	endpoint := os.Getenv("CRM_ENDPOINT")
	if endpoint == "" {
		return 0, nil, errors.New("missing CRM_ENDPOINT in environment")
	}

	url := endpoint + "/api/score/history/delete"
	method := "DELETE"

	// แปลง struct เป็น JSON
	reqBody := models.DeleteRequest{
		RefTransaction:   ref_transaction,
		RefTransactionID: ref_transaction_id,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	auth := os.Getenv("JOYLIDAY_AUTH")
	if auth == "" {
		return 0, nil, errors.New("missing JOYLIDAY_AUTH in environment")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+auth)

	res, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.StatusCode, nil, fmt.Errorf("request failed with status %s: %s", res.Status, string(body))
	}

	var resp models.DeleteHistoryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return res.StatusCode, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return res.StatusCode, &resp.Data, nil
}
func SyncScoreMember(input models.ScoreMember) (int, []byte, error) {
	endpoint := os.Getenv("CRM_ENDPOINT")
	if endpoint == "" {
		return 0, nil, errors.New("missing CRM_ENDPOINT in environment")
	}
	url := endpoint + "/api/customer/sync/score/member"

	method := "POST"

	// Marshal input to JSON
	jsonBody, err := json.Marshal(input)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	payload := bytes.NewReader(jsonBody)

	// Set timeout for HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	auth := os.Getenv("JOYLIDAY_AUTH")
	if auth == "" {
		return 0, nil, errors.New("missing JOYLIDAY_AUTH in environment")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+auth)

	// Execute request
	res, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	// Read response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle HTTP errors
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.StatusCode, body, fmt.Errorf("request failed with status %s", res.Status)
	}

	return res.StatusCode, body, nil
}
