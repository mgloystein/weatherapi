package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	weatherservice "github.com/mgloystein/weatherapp/services"
)

const (
	authUserAgent = "APITestGoWeatherClient/1.0"
	apiUrl        = "https://api.weather.gov/"
)

type WeatherCharacterizationResponse struct {
	Forecast         string `json:"forecast"`
	Characterization string `json:"characterization"`
}

func main() {
	mux := http.NewServeMux()
	service := weatherservice.NewWeatherService(http.DefaultClient, apiUrl)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		req := &weatherservice.WeatherRequest{}
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
		apiResp := &WeatherCharacterizationResponse{
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