package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var baseURL = "http://api.weatherapi.com/v1/"

type SupabaseData struct {
	Location       string
	Temperature    float64
	Humidity       float64
	IsDay          int8
	Wind_speed     float64
	Wind_direction string
	UV             float64
	Cloud          float64
}

type WeatherAPIData struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	Current struct {
		Temperature    float64 `json:"temp_c"`
		Humidity       float64 `json:"humidity"`
		IsDay          int8    `json:"is_day"`
		Wind_speed     float64 `json:"wind_kph"`
		Wind_direction string  `json:"wind_dir"`
		Uv             float64 `json:"uv"`
		Cloud          float64 `json:"cloud"`
	} `json:"current"`
}

func getWeatherData(apiKey, location, baseURL string) (WeatherAPIData, []byte, error) {
	url := fmt.Sprintf("%s/current.json?key=%s&q=%s", baseURL, apiKey, location)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("failed creating request", slog.Any("err", err))
		return WeatherAPIData{}, nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("client failed getting request", slog.Any("err", err))
		return WeatherAPIData{}, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed reading the response body", slog.Any("err ", err))
		return WeatherAPIData{}, nil, err
	}
	var weatherResponse WeatherAPIData
	err = json.Unmarshal(body, &weatherResponse)
	if err != nil {
		slog.Error("failed formatting the reponse into JSON", slog.Any("err ", err))
		return WeatherAPIData{}, nil, err
	}
	return weatherResponse, body, err
}

func formatWeatherData(apiResp WeatherAPIData) SupabaseData {
	return SupabaseData{
		Location:       apiResp.Location.Name,
		Temperature:    apiResp.Current.Temperature,
		Humidity:       apiResp.Current.Humidity,
		IsDay:          apiResp.Current.IsDay,
		Wind_speed:     apiResp.Current.Wind_speed,
		Wind_direction: apiResp.Current.Wind_direction,
		UV:             apiResp.Current.Uv,
		Cloud:          apiResp.Current.Cloud,
	}
}

func sendDataSupabase(apiKey, url string, formattedData SupabaseData) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	if cancel != nil {
		slog.Error("POST connection timmed out", slog.Any("err", cancel))
	}
	data, err := json.Marshal(formattedData)
	if err != nil {
		slog.Error("error marshalling the formatted weather data", slog.Any("err", err))
	}

	supaBaseURL := url + "/rest/v1/weather_data?select=*"
	req, err := http.NewRequestWithContext(ctx, "POST", supaBaseURL, bytes.NewReader(data))
	if err != nil {
		slog.Error("Request error", slog.Any("err", err))
	}

	req.Header = http.Header{
		"apikey":        []string{apiKey},
		"Authorizaiton": []string{"Bearer " + apiKey},
		"Content-Type":  []string{"application/json"},
		"Prefer":        []string{"return=representation"},
	}
	log.Println("Final JSON payload:", string(data))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("POST response failed", slog.Any("err", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("supabase returned status %d", resp.StatusCode)
	}
	return nil
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		slog.Error("failed loading .env file", slog.Any("err ", err))
	}
	weather, _, err := getWeatherData(os.Getenv("WEATHER_API_KEY"), "Montreal", os.Getenv("WEATHER_URL"))
	if err != nil {
		slog.Error("failed getting data", slog.Any("err ", err))
		return
	}
	formattedData := formatWeatherData(weather)
	err = sendDataSupabase(os.Getenv("SUPABASE_API"), os.Getenv("SUPABASE_URL"), formattedData)
	if err != nil {
		slog.Error("Supabase insertion failed", slog.Any("err", err))
		return
	}
	fmt.Println("Montreal weather is :", weather)
}
