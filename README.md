# Crawler Indexing App
This App will temporary save the crawler data and clean them, its gonna clean temp_item for mysql and add supervision data automatic base on supervision list

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


