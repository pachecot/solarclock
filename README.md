# Solar Clock

A Windows service and CLI toolkit for calculating solar event times (sunrise, sunset, solar noon) and sun position (elevation, azimuth) for any geographic location and date.

## Features

- REST API returning JSON or XML responses
- Runs as a native Windows service

## Commands

### `solarclock` — Windows Service

```
solarclock install [--port N]   Install as a Windows service (default port: 8080)
solarclock remove               Remove the Windows service
solarclock start                Start the service
solarclock stop                 Stop the service
solarclock status               get the service status
solarclock run [--port N]       Run in the foreground (debug mode)
```

The port is stored in the Windows Registry at:
`HKLM\SYSTEM\CurrentControlSet\Services\SolarClockService\Parameters`

## API

`solarclock` exposes the following endpoints:

### `GET /json`

Returns solar data as JSON.

### `GET /xml`

Returns solar data as XML.

### Query Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `lat`     | Yes      | Latitude in decimal degrees |
| `long`    | Yes      | Longitude in decimal degrees |
| `date`    | No       | Date in `YYYY-MM-DD` format (defaults to today) |
| `offset`  | No       | UTC timezone offset in hours (defaults to system timezone) |

### Example Request

```
GET http://localhost:8080/json?lat=42.21&long=-71.3&date=2026-03-14&offset=-5
```

### Example Response

```json
{
    "Day": "2026-03-14T05:00:00Z",
    "DateTime": "2026-03-14T15:20:00Z",
    "SunriseTime": "2026-03-14T10:47:00Z",
    "SunsetTime": "2026-03-14T22:53:00Z",
    "Offset": "-5h0m0s",
    "Date": "2026-03-14",
    "SolarNoon": "4:50PM",
    "Sunrise": "5:47AM",
    "Sunset": "5:53PM",
    "SunlightDuration": "12h5m...",
    "SolarElevation": 42.3,
    "SolarAzimuth": 185.7
}
```

## Building

```bash
# Windows service
go build -o solarclock.exe ./cmd/solarclock

# CLI tool
go build -o suntimes.exe ./cmd/suntimes

# Web UI server
go build -o scws.exe ./cmd/scws
```

## Dependencies

| Package | Purpose |
|---------|---------|
| [github.com/pachecot/solar](https://github.com/pachecot/solar) | Solar position calculations |
| [github.com/pachecot/julian](https://github.com/pachecot/julian) | Julian date conversion |
| [github.com/pachecot/angle](https://github.com/pachecot/angle) | Angle type and utilities |
| [golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys) | Windows service, registry, and event log APIs |

## Requirements

- Windows (service management uses Windows-only APIs)
- Go 1.24.1 or later

## License

MIT License — Copyright 2026 Tom Pacheco
