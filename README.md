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

Parametrs config with values by default:

```
GAME_LIST: "gameListPS_0.json"
GAME_DISSCOUNTS: "game_list_discounts.json"
GAME_SOURCE: "store.playstation.com"
GAME_PREFIX: "PS_"

BASE_URL_PS: "https://store.playstation.com"
BASE_URL_SALES_PS: "https://store.playstation.com/ru-ru/grid/STORE-MSF75508-PRICEDROPSCHI/"
BASE_URL_SALES_PARAM: "?gameContentType=games"

PROXY_HOST: ""
PROXY_PORT: 0
PROXY_USER: ""
PROXY_PASSWORD: ""

BOT_KEY: "your_bot_api_key"

CHANNEL_NAME: "@your_channel"

LOG_WARN: "logs.log"
```

## Run

To run the application, just execute a simple command

```
go run main.go
```

## Application Status

**[1.0.0] (2019-11-22) Warn:**
This application is currently under development. I do not recommend using it now.