A simple application to retrieve forecast data from the national weather service based on a give Geo location.

# Usage
To start run
```shell
go run main.go
```

# Requests
To test run
```shell
curl --location 'http://localhost:8080/weather' \
--header 'Content-Type: application/json' \
--header 'Accept: application/json' \
--data '{
    "longitude": -104.0420525,
    "latitude":37.8w17411
}'
```