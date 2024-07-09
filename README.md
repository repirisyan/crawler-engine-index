# Crawler Indexing App
This app will indexing data from new crawler data and store the data to nosql database (mongodb), when the data is stored inside mongodb, it will find and delete the duplicate data base on (title, seller and marketplace)

## System Requirement
- NodeJS 20.x (always using the LTS version)
- NPM 10.x

## Instalation
- go get -u all

## Development Env
- go run main.go

## Production Env
- go build -o crawler-index
- move the file to crawler-engine-indexing inside folder goapp
- the execution file will read .env from parent app


