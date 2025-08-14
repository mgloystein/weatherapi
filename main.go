package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/mgloystein/weatherapp/serives"
)

const (
	authUserAgent = "APITestGoWeatherClient/1.0"
	apiUrl        = "https://api.weather.gov/"
)

func main() {
	mux := http.NewServeMux()
	service := serives.NewWeatherService(http.DefaultClient, apiUrl)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		var req *WeatherRequest
		ctx := context.Background()
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := service.GetCurrentWeather(ctx, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(resp.Periods) == 0 {
			http.Error(w, "no weather data available", http.StatusInternalServerError)
		}

		period := resp.Periods[0]
		apiResp := &TestWeatherServiceResponse{
			Forecast: period.Forecast,
		}

		if period.Temperature <= 32 {
			apiResp.Characterization = "cold"
		} else if period.Temperature <= 68 {
			apiResp.Characterization = "moderate"
		} else {
			apiResp.Characterization = "hot"
		}

		// fmt.Println("Weather Response:", resp)
		data, err := json.Marshal(apiResp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Start the server in a separate goroutine so it doesn't block.
	go func() {
		fmt.Println("Starting server on http://localhost:8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			// If an error occurs, log it and let the main function know.
			fmt.Printf("Could not start server: %v\n", err)
		}
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)
}

type TestWeatherServiceResponse struct {
	Forecast         string `json:"forecast"`
	Characterization string `json:"characterization"`
}

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
