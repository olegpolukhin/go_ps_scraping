# Scraping Discounts Bot

The parse discounted games from PlayStation Store and post this games to telegram channel.

## Install

Install package

```
go get -v github.com/olegpolukhin/go_ps_scraping
```

Dependency Installation

```
go get -v
```

## Config

You can change the contents of the config file `config.yml`

Parametrs config:
- GAME_LIST
- GAME_DISSCOUNTS
- GAME_SOURCE
- GAME_PREFIX
- BOT_KEY
- LOG_WARN
- CHANNEL_NAME
- BASE_URL_PS
- BASE_URL_SALES_PS
- BASE_URL_SALES_PARAM

## Run

To run the application, just execute a simple command

```
go run main.go
```

## Application Status

**[1.0.0] (2019-11-22) Warn:**
This application is currently under development. I do not recommend using it now.