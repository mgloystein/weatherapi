package service

import (
	"context"
	"encoding/json"
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
	forecastURL, err := ws.getPoint(ctx, req.Latitude, req.Longitude)
	if err != nil {
		return nil, err
	}
	// fmt.Println("Forecast URL:", forecastURL)
	apiReq, err := http.NewRequestWithContext(ctx, "GET", forecastURL, nil)
	if err != nil {
		return nil, err
	}
	apiReq.Header.Set("Accept", "application/ld+json")
	apiReq.Header.Set("User-Agent", "APITestGoWeatherClient/1.0")

	apiResp, err := ws.client.Do(apiReq)
	if err != nil {
		return nil, err
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get point data: status %d", apiResp.StatusCode)
	}
	resp = &WeatherResponse{}
	err = json.NewDecoder(apiResp.Body).Decode(&resp)
	return
}

func (ws *weatherService) getPoint(ctx context.Context, lat, lon float64) (string, error) {
	url := fmt.Sprintf("%spoints/%.4f,%.4f", ws.baseURL, lat, lon)
	// fmt.Println("Point URL:", url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/ld+json")
	req.Header.Set("User-Agent", "APITestGoWeatherClient/1.0")

	resp, err := ws.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get point data: status %d", resp.StatusCode)
	}

	var result struct {
		Forecast string `json:"forecast"`
	}
	data := []byte{}
	data, err = io.ReadAll(resp.Body)
	// fmt.Println("Raw Point Response:", string(data))
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", err
	}
	// fmt.Println("Decoded Point Response:", result)
	return result.Forecast, nil
}
