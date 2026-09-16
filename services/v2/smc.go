package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"new-pos-api/config"
	model_v2 "new-pos-api/models/v2"
	"os"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	smcClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	tokenCache     string
	tokenCacheLock sync.RWMutex
	lastLoginTime  time.Time
)

// LoginSMC returns the access token from the SMC API with multiple cache layers
func LoginSMC() (string, error) {
	// 1. FAST PATH: Local Memory Cache (Thread-safe)
	tokenCacheLock.RLock()
	// Usually tokens are valid for 50-60m, we check for 45m to be safe
	if tokenCache != "" && time.Since(lastLoginTime) < 45*time.Minute {
		defer tokenCacheLock.RUnlock()
		return tokenCache, nil
	}
	tokenCacheLock.RUnlock()

	// 2. Slow Path: Acquire Lock to handle refresh
	tokenCacheLock.Lock()
	defer tokenCacheLock.Unlock()

	// Double check memory (another goroutine might have refreshed it)
	if tokenCache != "" && time.Since(lastLoginTime) < 45*time.Minute {
		return tokenCache, nil
	}

	// 3. Proactively deactivate expired tokens in DB
	DeactivateToken()

	// 4. Shared Cache: Check DB for active token (from another instance)
	activeToken, err := GetValidTokenFromDB()
	if err == nil && activeToken != "" {
		tokenCache = activeToken
		lastLoginTime = time.Now()
		return activeToken, nil
	}

	// 4. API REFRESH: Both caches failed
	endpoint := os.Getenv("SMC_API_ENDPOINT")
	if endpoint == "" {
		return "", fmt.Errorf("SMC_API_ENDPOINT is not set")
	}

	url := fmt.Sprintf("%s/v1/login", endpoint)
	// fmt.Printf("url: %s\n", url)
	loginReq := model_v2.SMCLoginRequest{
		Username: os.Getenv("SMC_USR"),
		Password: os.Getenv("SMC_PWD"),
	}

	jsonBody, err := json.Marshal(loginReq)
	if err != nil {
		return "", fmt.Errorf("marshal login request: %w", err)
	}
	// fmt.Printf("jsonBody: %s\n", string(jsonBody))
	resp, err := smcClient.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	// fmt.Printf("resp: %v\n", resp)
	// fmt.Printf("err: %v\n", err)
	if err != nil {
		return "", fmt.Errorf("send login request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("SMC login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginRes model_v2.SMCLoginResponse
	if err := json.Unmarshal(body, &loginRes); err != nil {
		return "", fmt.Errorf("unmarshal login response: %w", err)
	}

	// Update Memory Cache
	tokenCache = loginRes.Token
	lastLoginTime = time.Now()

	// Update DB (persist for other instances)
	InsertToken(tokenCache)

	return tokenCache, nil
}

// SendPrizeSMC calls the prize endpoint on the SMC API
func SendPrizeSMC(prizeReq model_v2.SMCPrizeDto) (*model_v2.SMCPrizeResponse, error) {
	endpoint := os.Getenv("SMC_API_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("SMC_API_ENDPOINT is not set")
	}
	token, errToken := LoginSMC()
	if errToken != nil {
		return nil, fmt.Errorf("Can not Authen Smart crane")
	}

	url := fmt.Sprintf("%s/v1/prize", endpoint)
	jsonBody, err := json.Marshal(prizeReq)
	if err != nil {
		return nil, fmt.Errorf("marshal prize request: %w", err)
	}

	client := smcClient
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create prize request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send prize request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read prize response: %w", err)
	}
	// fmt.Printf("SMC Prize Response xxx: %s\n", string(body))

	var prizeRes model_v2.SMCPrizeResponse
	if err := json.Unmarshal(body, &prizeRes); err != nil {
		return nil, fmt.Errorf("unmarshal prize response: %w", err)
	}

	if !prizeRes.Success {
		return &prizeRes, fmt.Errorf("SMC prize failed: %s", string(body))
	}	
	return &prizeRes, nil
}
func InsertToken(token string) {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	startTime := time.Now().In(loc)
	endTime := startTime.Add(50 * time.Minute)
	db := config.DB_JREADER

	// Use Transaction to ensure old tokens are deactivated before creating new one
	db.Transaction(func(tx *gorm.DB) error {
		tx.Model(&model_v2.SmcTk{}).Where("is_active = ?", true).Update("is_active", false)
		smcTk := model_v2.SmcTk{
			Token:     token,
			StartDate: startTime,
			EndDate:   &endTime,
			IsActive:  true,
		}
		return tx.Create(&smcTk).Error
	})
}

func GetValidTokenFromDB() (string, error) {
	db := config.DB_JREADER
	var smcTk model_v2.SmcTk
	loc, _ := time.LoadLocation("Asia/Bangkok")
	now := time.Now().In(loc)

	// Single query: IsActive AND not expired
	result := db.Where("is_active = ? AND end_date > ?", true, now).Order("row_no DESC").First(&smcTk)
	// fmt.Printf("result: %v\n", result)
	if result.Error != nil {
		return "", result.Error
	}
	return smcTk.Token, nil
}

func DeactivateToken() error {
	ctx := context.Background()
	now, _ := time.LoadLocation("Asia/Bangkok")
	endTime := time.Now().In(now)
	db := config.DB_JREADER

	result := db.WithContext(ctx).Model(&model_v2.SmcTk{}).
		Where("is_active = ? AND end_date <= ?", true, endTime).
		Update("is_active", false)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		fmt.Printf("Deactivated %d expired SMC tokens\n", result.RowsAffected)
	}

	return nil
}
