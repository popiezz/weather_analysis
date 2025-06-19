# Weather Data Uploader

This Go application periodically fetches current weather data from a weather API and stores it in a Supabase table. It also exposes an HTTP endpoint to manually trigger this data upload.

## Features

- Periodic weather data collection (default: every 15 minutes)
- Automatic upload to a Supabase backend
- Manual trigger via HTTP POST endpoint
- Secure configuration using `.env` file

## Project Structure

```
weather/
├── main.go
├── .env
├── go.mod
├── go.sum
```

## Environment Variables

Create a `.env` file in the root of the project with the following keys:

```env
SUPABASE_API_KEY=your_supabase_api_key
SUPABASE_URL=https://your-project.supabase.co
WEATHER_API_KEY=your_weatherapi_key
WEATHER_URL=http://api.weatherapi.com/v1/current.json?q=YourLocation
```

- `SUPABASE_API_KEY`: Your Supabase service key
- `SUPABASE_URL`: The URL of your Supabase project (e.g. `https://xyzcompany.supabase.co`)
- `WEATHER_API_KEY`: API key from weather provider (e.g. [weatherapi.com](https://www.weatherapi.com))
- `WEATHER_URL`: Full request URL to get weather data for your location

## Supabase Request Payload

This is the structure of the JSON data sent to Supabase on each request:

```json
{
  "timestamp": "2025-06-19T00:32:34Z",
  "location": "Montreal",
  "temperature": 29.1,
  "humidity": 58,
  "wind_speed": 9,
  "conditions": "Moderate or heavy rain shower",
  "wind_degrees": 0,
  "wind_dir": "SW",
  "pressure_mb": 1008,
  "feelslike_c": 30.3,
  "uv": 6,
  "vis_km": 10.0,
  "cloud": 75
}
```

Make sure your Supabase table schema matches these keys and types.

## How It Works

1. Loads environment variables from `.env`.
2. Starts a ticker that triggers every 15 minutes.
3. On each tick, it:
   - Fetches weather data from the configured API.
   - Formats and sends the data to a Supabase table using HTTP POST.
4. Exposes an HTTP endpoint at `POST /update` to manually trigger the same process.

## Run the Project

### Prerequisites

- Go 1.18+
- Supabase account and configured table
- A weather API key (from WeatherAPI or similar)

### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/weather-uploader.git
   cd weather-uploader
   ```

2. Create a `.env` file using the format above.

3. Run the application:
   ```bash
   go run main.go
   ```

You will see logs every 15 minutes indicating a successful data upload or error.

## Manual Update Trigger

You can manually trigger a data upload using:

```bash
curl -X POST http://localhost:8080/update
```

## Configuration

To change the interval (default: 15 minutes), modify this line in `main.go`:

```go
interval := 15 * time.Minute
```

## TODO

- Add retry logic on Supabase failures
- Store logs in a persistent file or service
- Add unit tests and CI
- Dockerize the project

## License

This project is open-sourced under the MIT License. See `LICENSE` file for details.
