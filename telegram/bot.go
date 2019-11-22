package telegram

import (
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/olegpolukhin/go_ps_scraping/models"

	telegramApi "github.com/go-telegram-bot-api/telegram-bot-api"
	datasource "github.com/olegpolukhin/go_ps_scraping/datasource"
	taskmanager "github.com/olegpolukhin/go_ps_scraping/taskmanager"
)

var bot *telegramApi.BotAPI

type postGameDiscount struct {
	HeaderTitle    string
	GameTitle      string
	DiscountString string
	PriceString    string
	GameCoverURL   string
	GameURL        string
}

func generatePostFromSource() (newPost postGameDiscount, screenshots []string) {
	var gameToPost models.GameGeneral
	var priceString = ""

	gameToPost = datasource.GetRandomDiscountedGame()
	gameTitle := "PlayStation Store, Скидки на сегодня"
	gameName := gameToPost.Name
	if strconv.FormatInt(gameToPost.Price, 10) == "0" {
		priceString = "Перейдите по ссылке, чтобы узнать итоговую цену"
	} else {
		priceString = strconv.FormatInt(gameToPost.Price, 10)
	}

	if gameToPost.IsFree {
		newPost = postGameDiscount{
			gameTitle,
			gameName,
			"",
			"",
			gameToPost.HeaderImageURL,
			gameToPost.Link,
		}
	} else {
		newPost = postGameDiscount{
			gameTitle,
			gameName,
			strconv.FormatInt(gameToPost.Discount, 10),
			priceString,
			gameToPost.HeaderImageURL,
			gameToPost.Link,
		}
	}

	return
}

func genegateBundlePostFromSource(fromSourceType string, bundleSize int) (gamePostBundle []postGameDiscount, gamePostBundleCovers []string) {
	counter := 0
	for counter < bundleSize {
		gamePost, _ := generatePostFromSource()
		if gamePost.GameTitle == "" {
			return
		}
		gamePostBundle = append(gamePostBundle, gamePost)
		gamePostBundleCovers = append(gamePostBundleCovers, gamePost.GameCoverURL)
		counter++

	}
	return gamePostBundle, gamePostBundleCovers
}

// BotServerProcess service by sent to channel telegramm
func BotServerProcess(inKey string, controlChannel chan string) {
	var err error
	bot, err = telegramApi.NewBotAPI(inKey)
	if err != nil {
		logrus.Panic("BotServerProcess NewBotAPI error: ", err)
	}

	var mBot *telegramApi.BotAPI

	mBot = bot
	mBot.Debug = false

	botAuthMsg := fmt.Sprintf("Telegram bot %s re/started, Time - %s", mBot.Self.UserName, time.Now().String())

	logrus.Info(fmt.Sprintf("Authorized as %s", mBot.Self.UserName))
	logrus.Info(botAuthMsg)

	if mBot.Debug {
		msg := telegramApi.NewMessageToChannel(viper.GetString("CHANNEL_NAME"), botAuthMsg)
		bot.Send(msg)
	}
}

// BotWaitProcess .
func BotWaitProcess(inKey string) {
	var err error
	bot, err = telegramApi.NewBotAPI(inKey)
	if err != nil {
		logrus.Panic("BotWaitProcess NewBotAPI error: ", err)
	}

	var mBot *telegramApi.BotAPI
	mBot = bot
	mBot.Debug = false

	logrus.Info(fmt.Sprintf("Bot Wait Init. Authorized as %s", mBot.Self.UserName))

	if mBot.Debug {
		botAuthMsg := fmt.Sprintf("Telegram bot %s re/started, Time - %s", mBot.Self.UserName, time.Now().String())
		msg := telegramApi.NewMessageToChannel(viper.GetString("CHANNEL_NAME"), botAuthMsg)
		bot.Send(msg)
	}

	var messageChannel = telegramApi.NewUpdate(0)
	messageChannel.Timeout = 60

	updates, err := mBot.GetUpdatesChan(messageChannel)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		messageText := update.Message.Text

		switch messageText {
		case "sendTest":
			msg := telegramApi.NewMessageToChannel(viper.GetString("CHANNEL_NAME"), "Empty test message")
			bot.Send(msg)
		case "game_ps":
			somePost, _ := generatePostFromSource()
			msgString := somePost.HeaderTitle
			msgString += "\n" + somePost.GameTitle
			msgString += "\nСкидка: " + somePost.DiscountString + "% \nЦена: " + somePost.PriceString + " руб."
			msgString += "\nСсылка: " + somePost.GameURL

			msgMain := telegramApi.NewMessageToChannel(viper.GetString("CHANNEL_NAME"), msgString)
			bot.Send(msgMain)
			// var chatID int64 = -77777777777
			// file.DownloadImage(somePost.GameCoverURL, "cover_ps.jpg", func() {
			// 	msgCover := telegramApi.NewPhotoUpload(chatID, "cover_ps.jpg")
			// 	bot.Send(msgCover)
			// 	bot.Send(msgMain)
			// })
		case "game_bundle_ps":
			task := SendPostGameBundleTask(3)
			task()
		default:
			continue
		}
	}
}

// SendPostGameBundleTask send game Bundle
func SendPostGameBundleTask(bundleSize int) taskmanager.SingleTask {
	return func() {
		// var uploadedCovers []string
		var postMessages []string
		gamePostBundle, _ := genegateBundlePostFromSource("ps", bundleSize)

		if len(gamePostBundle) == 0 {
			logrus.Warn("Game Bundle list is Empty!")
			return
		}

		for _, gamePost := range gamePostBundle {
			msgString := "\n" + gamePost.GameTitle
			msgString += "\nСкидка: " + gamePost.DiscountString + "% \nЦена: " + gamePost.PriceString + " руб."
			msgString += "\nСсылка: " + gamePost.GameURL
			postMessages = append(postMessages, msgString)
			// file.DownloadImage(gamePost.GameCoverURL, "cover_ps_bundle_"+strconv.Itoa(index)+".jpg", func() {
			// 	uploadedCovers = append(uploadedCovers, "cover_ps_bundle_"+strconv.Itoa(index)+".jpg")
			// })
		}

		msgBundleHeader := telegramApi.NewMessageToChannel(viper.GetString("CHANNEL_NAME"), "Бандл скидок PlayStation Store к этому часу!")
		bot.Send(msgBundleHeader)

		for _, message := range postMessages {
			msgMain := telegramApi.NewMessageToChannel(viper.GetString("CHANNEL_NAME"), message)
			// var chatID int64 = -77777777777
			// msgCover := telegramApi.NewPhotoUpload(chatID, uploadedCovers[index])
			// bot.Send(msgCover)
			bot.Send(msgMain)
		}
	}
}
