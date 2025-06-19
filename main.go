package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type weatherData struct {
	Timestamp   string  `json:"timestamp"`
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Wind_speed  float64 `json:"wind_speed"`
	Conditions  string  `json:"conditions"`
	Wind_degree float64 `json:"wind_degree"`
	Wind_dir    string  `json:"wind_dir"`
	Pressure_mb float64 `json:"pressure_mb"`
	Cloud       float64 `json:"cloud"`
	Feelslike_c float64 `json:"feelslike_c"`
	Vis_km      float64 `json:"vis_km"`
	Uv          float64 `json:"uv"`
	Is_day      float64 `json:"is_day"`
}

type WeatherAPIResponse struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		Humidity  float64 `json:"humidity"`
		WindKph   float64 `json:"wind_kph"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
		LastUpdated string  `json:"last_updated"`
		Wind_degree float64 `json:"wind_degree"`
		Wind_dir    string  `json:"wind_dir"`
		Pressure_mb float64 `json:"pressure_mb"`
		Cloud       float64 `json:"cloud"`
		Feelslike_c float64 `json:"feelslike_c"`
		Vis_km      float64 `json:"vis_km"`
		Uv          float64 `json:"uv"`
	} `json:"current"`
}

func sendWeatherDataCore(apiKey, supabaseURL, weatherApiKey, weatherURL string) error {
	weatherResp, bodyBytes, err := getWeatherData(weatherApiKey, weatherURL)
	if err != nil {
		return fmt.Errorf("failed to get weather data: %w", err)
	}

	formatted := formatWeatherForSupabase(weatherResp)
	// weather := []weatherData{formatted}

	data, err := json.Marshal(formatted)
	if err != nil {
		return fmt.Errorf("failed to marshal weather data: %w", err)
	}

	url := supabaseURL + "/rest/v1/weather_data?select=*"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create Supabase request: %w", err)
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	log.Println("Final JSON payload:", string(data))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Supabase: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("supabase returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func sendWeatherData(apiKey, supabaseURL, weatherApiKey, weatherURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := sendWeatherDataCore(apiKey, supabaseURL, weatherApiKey, weatherURL)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Weather data sent succesfully"})
	}
}

func getWeatherData(apiKey, weatherURL string) (WeatherAPIResponse, []byte, error) {
	url := fmt.Sprintf("%s/current.json?key=%s&q=Montreal", weatherURL, apiKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return WeatherAPIResponse{}, nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return WeatherAPIResponse{}, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return WeatherAPIResponse{}, nil, err
	}
	var weatherResp WeatherAPIResponse
	err = json.Unmarshal(body, &weatherResp)
	if err != nil {
		return WeatherAPIResponse{}, nil, err
	}
	return weatherResp, body, nil
}

func formatWeatherForSupabase(apiResp WeatherAPIResponse) weatherData {
	return weatherData{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Location:    apiResp.Location.Name,
		Temperature: apiResp.Current.TempC,
		Humidity:    apiResp.Current.Humidity,
		Wind_speed:  apiResp.Current.WindKph,
		Conditions:  apiResp.Current.Condition.Text,
		Wind_degree: apiResp.Current.Wind_degree,
		Wind_dir:    apiResp.Current.Wind_dir,
		Pressure_mb: apiResp.Current.Pressure_mb,
		Cloud:       apiResp.Current.Cloud,
		Feelslike_c: apiResp.Current.Feelslike_c,
		Vis_km:      apiResp.Current.Vis_km,
		Uv:          apiResp.Current.Uv,
	}
}

func main() {
	// LOAD .env FILES
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	api_key := os.Getenv("SUPABASE_API_KEY")
	supabaseURL := os.Getenv("SUPABASE_URL")
	weather_api := os.Getenv("WEATHER_API_KEY")
	weatherURL := os.Getenv("WEATHER_URL")

	// CREATE TICKER AND ALL ITS DATA
	interval := 15 * time.Minute
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			err := sendWeatherDataCore(api_key, supabaseURL, weather_api, weatherURL)
			if err != nil {
				log.Println("Ticker error:", err)
			} else {
				log.Println("Weather data sent successfully via ticker")
			}
		}
	}()

	router := gin.Default()
	router.POST("/update", sendWeatherData(api_key, supabaseURL, weather_api, weatherURL))
	router.Run("localhost:8080")
}
