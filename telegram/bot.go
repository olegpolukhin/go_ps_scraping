package telegram

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/olegpolukhin/go_ps_scraping/file"
	"github.com/olegpolukhin/go_ps_scraping/proxy"

	telegramApi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/olegpolukhin/go_ps_scraping/datasource"
	taskmanager "github.com/olegpolukhin/go_ps_scraping/taskmanager"
)

var bot *telegramApi.BotAPI

func initBot(inKey string) (bot *telegramApi.BotAPI, err error) {
	if viper.GetString("PROXY_USER") == "" || viper.GetString("PROXY_PASSWORD") == "" {
		bot, err = telegramApi.NewBotAPI(inKey)
	} else {
		tr := proxy.NewProxyTransport()
		bot, err = telegramApi.NewBotAPIWithClient(inKey, &http.Client{
			Transport: tr,
		})
	}

	return
}

// BotServerProcess service by sent to channel telegramm
func BotServerProcess(inKey string) {
	var err error
	var mBot *telegramApi.BotAPI

	bot, err = initBot(inKey)
	if err != nil {
		logrus.Panic("BotServerProcess NewBotAPI error: ", err)
	}

	mBot = bot
	mBot.Debug = false

	botAuthMsg := fmt.Sprintf("Telegram bot %s re/started, Time - %s", mBot.Self.UserName, time.Now().String())

	logrus.Info(fmt.Sprintf("Authorized as %s", mBot.Self.UserName))
	logrus.Info(botAuthMsg)

	if mBot.Debug {
		msg := telegramApi.NewMessageToChannel(viper.GetString("CHANNEL_NAME"), botAuthMsg)
		bot.Send(msg)
	}

	var messageChannel = telegramApi.NewUpdate(0)
	messageChannel.Timeout = 60

	updates, err := mBot.GetUpdatesChan(messageChannel)
	if err != nil {
		logrus.Panic("Bot GetUpdatesChan error: ", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		messageText := update.Message.Text

		switch messageText {
		case "sendTest":
			msg := telegramApi.NewMessageToChannel(viper.GetString("CHANNEL_NAME"), "Empty test message")
			if _, err = bot.Send(msg); err != nil {
				logrus.Panic("Bot send message error: ", err)
			}
		case "game_ps":

			var uploadedCover string
			var msgConfig telegramApi.Message
			var chatID int64

			msgInit := telegramApi.NewMessageToChannel(viper.GetString("CHANNEL_NAME"), "Отличная скидка в PlayStation Store!")
			msgConfig, err = bot.Send(msgInit)
			if err != nil {
				logrus.Panic("Bot send message error: ", err)
			}

			somePost, _ := datasource.GeneratePostFromSource()
			msgString := somePost.HeaderTitle
			msgString += "\n" + somePost.GameTitle
			msgString += "\nСкидка: " + somePost.DiscountString + "% \nЦена: " + somePost.PriceString + " руб."
			msgString += "\nСсылка: " + somePost.GameURL

			msgMain := telegramApi.NewMessageToChannel(viper.GetString("CHANNEL_NAME"), msgString)

			uploadedCover = "cover_game.jpg"
			file.DownloadImage(somePost.GameCoverURL, uploadedCover, func() {
				if msgConfig.Chat != nil {
					chatID = msgConfig.Chat.ID
					msgCover := telegramApi.NewPhotoUpload(chatID, uploadedCover)
					if _, err = bot.Send(msgCover); err != nil {
						logrus.Panic("Bot send cover message error: ", err)
					}
				}
			})

			if _, err = bot.Send(msgMain); err != nil {
				logrus.Panic("Bot send message error: ", err)
			}

		case "game_bundle_ps":
			task := SendPostGameBundleTask(viper.GetInt("GAME_BUNDLE"))
			task()
		default:
			continue
		}
	}
}

// SendPostGameBundleTask send game Bundle
func SendPostGameBundleTask(bundleSize int) taskmanager.SingleTask {
	return func() {
		var chatID int64
		var uploadedCovers []string
		var postMessages []string
		gamePostBundle, _ := datasource.GenegateBundlePostFromSource(bundleSize)

		if len(gamePostBundle) == 0 {
			logrus.Warn("Game Bundle list is Empty!")
			return
		}

		for index, gamePost := range gamePostBundle {
			msgString := "\n" + gamePost.GameTitle
			msgString += "\nСкидка: " + gamePost.DiscountString + "% \nЦена: " + gamePost.PriceString + " руб."
			msgString += "\nСсылка: " + gamePost.GameURL
			postMessages = append(postMessages, msgString)
			file.DownloadImage(gamePost.GameCoverURL, "cover_game_"+strconv.Itoa(index)+".jpg", func() {
				uploadedCovers = append(uploadedCovers, "cover_game_"+strconv.Itoa(index)+".jpg")
			})
		}

		msgBundleHeader := telegramApi.NewMessageToChannel(viper.GetString("CHANNEL_NAME"), "Бандл скидок PlayStation Store к этому часу!")

		msgConfig, err := bot.Send(msgBundleHeader)
		if err != nil {
			logrus.Warn("Bot send message error: ", err)
			return
		}

		for index, message := range postMessages {
			msgMain := telegramApi.NewMessageToChannel(viper.GetString("CHANNEL_NAME"), message)

			if msgConfig.Chat != nil {
				chatID = msgConfig.Chat.ID
				msgCover := telegramApi.NewPhotoUpload(chatID, uploadedCovers[index])
				bot.Send(msgCover)
			}

			bot.Send(msgMain)
		}
	}
}
