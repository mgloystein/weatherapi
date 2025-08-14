package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type WeatherRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type WeatherResponse struct {
	Properties struct {
		Units string `json:"units"`
	} `json:"properties"`
	Periods []struct {
		Forecast    string `json:"shortForecast"`
		Temperature int32  `json:"temperature"`
	}
}

type WeatherService interface {
	GetCurrentWeather(ctx context.Context, req *WeatherRequest) (resp *WeatherResponse, err error)
}

type weatherService struct {
	client  *http.Client
	baseURL string
}

func NewWeatherService(client *http.Client, baseURL string) WeatherService {
	return &weatherService{
		client:  client,
		baseURL: baseURL,
	}
}

func (ws *weatherService) GetCurrentWeather(ctx context.Context, req *WeatherRequest) (resp *WeatherResponse, err error) {
	forecastURL, err := ws.getPointURL(ctx, req.Latitude, req.Longitude)
	if err != nil {
		return nil, err
	}
	// fmt.Println("Forecast URL:", forecastURL)
	apiResp, err := ws.makeRequest(ctx, forecastURL)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to make API request"))
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get point data: status %d", apiResp.StatusCode)
	}
	resp = &WeatherResponse{}
	err = json.NewDecoder(apiResp.Body).Decode(&resp)
	return
}

func (ws *weatherService) getPointURL(ctx context.Context, lat, lon float64) (string, error) {
	url := fmt.Sprintf("%spoints/%.4f,%.4f", ws.baseURL, lat, lon)
	// fmt.Println("Point URL:", url)
	apiResp, err := ws.makeRequest(ctx, url)
	if err != nil {
		return "", errors.Join(err, errors.New("failed to make API request"))
	}

	var result struct {
		Forecast string `json:"forecast"`
	}
	var data []byte
	data, err = io.ReadAll(apiResp.Body)
	if err != nil {
		return "", errors.Join(err, errors.New("failed to read API response"))
	}
	// fmt.Println("Raw Point Response:", string(data))
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", err
	}
	// fmt.Println("Decoded Point Response:", result)
	return result.Forecast, nil
}

func (ws *weatherService) makeRequest(ctx context.Context, url string) (*http.Response, error) {
	apiReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	apiReq.Header.Set("Accept", "application/ld+json")
	apiReq.Header.Set("User-Agent", "APITestGoWeatherClient/1.0")

	apiResp, err := ws.client.Do(apiReq)
	if err != nil {
		return nil, err
	}
	return apiResp, nil
}
