package telegram

import (
	"fmt"
	"log"
	"github.com/olegpolukhin/go_ps_scraping/logger"
	"github.com/olegpolukhin/go_ps_scraping/models"
	"strconv"
	"time"

	datasource "github.com/olegpolukhin/go_ps_scraping/datasource"
	"github.com/olegpolukhin/go_ps_scraping/file"
	taskmanager "github.com/olegpolukhin/go_ps_scraping/taskmanager"

	telegramApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var mBot *telegramApi.BotAPI
var KEY = "1043462173:AAGVqh6R0Nk0-kW0mQAX5k8JiT8Rib-OY7g"
var minimumMetacriticScore int64 = 0
var postingPeriod int64
var postingPeriodType = "hour"
var bot *telegramApi.BotAPI

var TESTING_CHANNEL_NAME = "@game_demo_ps"
var TESTING_CHANNEL_CHAT_ID int64 = -77777777777

var CHANNEL_PS_NAME = "@game_demo_ps"
var CHANNEL_CHAT_ID int64 = -77777777777

var POSTING_START_HOUR = 10
var POSTING_END_HOUR = 19

var steamPostingStarted = false
var gogPostingStarted = false
var psPostingStarted = false

type PostGameDiscount struct {
	HeaderTitle    string
	GameTitle      string
	DiscountString string
	PriceString    string
	GameCoverURL   string
	GameURL        string
}

func GeneratePostFromSource(fromSourceType string) (newPost PostGameDiscount, screenshots []string) {
	var gameToPost models.GameGeneral
	var gameTitle = ""
	var gameName = ""
	var priceString = ""

	switch fromSourceType {
	case "ps":
		gameToPost = datasource.PsGetRandomDiscountedGame()
		gameTitle = "PlayStation Store, Скидки на сегодня"
		gameName = gameToPost.Name
		if strconv.FormatInt(gameToPost.Price, 10) == "0" {
			priceString = "Перейдите по ссылке, чтобы узнать итоговую цену"
		} else {
			priceString = strconv.FormatInt(gameToPost.Price, 10)
		}
	}

	if gameToPost.IsFree {
		newPost = PostGameDiscount{
			gameTitle,
			gameName,
			"",
			"",
			gameToPost.HeaderImageURL,
			gameToPost.Link,
		}
	} else {
		newPost = PostGameDiscount{
			gameTitle,
			gameName,
			strconv.FormatInt(gameToPost.Discount, 10),
			priceString,
			gameToPost.HeaderImageURL,
			gameToPost.Link,
		}
	}

	return newPost, screenshots
}

func genegateBundlePostFromSource(fromSourceType string, bundleSize int) (gamePostBundle []PostGameDiscount, gamePostBundleCovers []string) {
	counter := 0
	for counter < bundleSize {
		gamePost, _ := GeneratePostFromSource("ps")
		gamePostBundle = append(gamePostBundle, gamePost)
		gamePostBundleCovers = append(gamePostBundleCovers, gamePost.GameCoverURL)
		counter++
	}
	return gamePostBundle, gamePostBundleCovers
}

// BotServerProcess service by sent channel
func BotServerProcess(inKey string, controlChannel chan string) {
	var err error
	bot, err = telegramApi.NewBotAPI(inKey)
	if err != nil {
		log.Panic("NewBotAPI", err)
	}
	mBot = bot
	mBot.Debug = true

	taskControlChannel := make(chan string)

	log.Printf("Authorized as %s", mBot.Self.UserName)
	fmt.Println("Telegram bot " + mBot.Self.UserName + " re/started, Time - " + time.Now().String())
	msg := telegramApi.NewMessageToChannel(TESTING_CHANNEL_NAME, "Telegram bot "+mBot.Self.UserName+" re/started "+time.Now().String()+"\nAll configs lost")
	bot.Send(msg)

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
			msg := telegramApi.NewMessageToChannel(TESTING_CHANNEL_NAME, "Empty test message")
			bot.Send(msg)
		case "game_ps":
			somePost, _ := GeneratePostFromSource("ps")
			msgString := somePost.HeaderTitle
			msgString += "\n" + somePost.GameTitle
			msgString += "\nСкидка: " + somePost.DiscountString + "% \nЦена: " + somePost.PriceString + " руб."
			msgString += "\nСсылка: " + somePost.GameURL

			msgMain := telegramApi.NewMessageToChannel(CHANNEL_PS_NAME, msgString)
			bot.Send(msgMain)

			// file.DownloadImage(somePost.GameCoverURL, "cover_ps.jpg", func() {
			// 	msgCover := telegramApi.NewPhotoUpload(CHANNEL_CHAT_ID, "cover_ps.jpg")
			// 	bot.Send(msgCover)
			// 	bot.Send(msgMain)
			// })
		case "game_bundle_ps":
			task := GetPsPostGameBundleTask(4)
			task()
		case "stop_posting":
			taskControlChannel <- "stop"
		case "barguzin":
			controlChannel <- "end"
		}
	}
}

func GetPsPostingPeriodicTask(taskControlChannel chan string) taskmanager.SingleTask {
	return func() {
		if !psPostingStarted {
			go taskmanager.StartPeriodicTask(2, "hour", POSTING_START_HOUR, POSTING_END_HOUR, taskControlChannel, func() {
				somePost, _ := GeneratePostFromSource("ps")
				msgString := somePost.HeaderTitle
				msgString += "\n" + somePost.GameTitle
				msgString += "\nСкидка: " + somePost.DiscountString + "% \nЦена: " + somePost.PriceString + " руб."
				msgString += "\nСсылка: " + somePost.GameURL
				msgMain := telegramApi.NewMessageToChannel(CHANNEL_PS_NAME, msgString)
				if somePost.HeaderTitle != "" {
					file.DownloadImage(somePost.GameCoverURL, "cover_ps.jpg", func() {
						msgCover := telegramApi.NewPhotoUpload(CHANNEL_CHAT_ID, "cover_ps.jpg")
						bot.Send(msgCover)
						bot.Send(msgMain)
					})
				} else {
					bot.Send(msgMain)
				}
			})
			psPostingStarted = true
			logger.Write("PsPostingPeriodicTask - ps posting initial started")
		} else {
			logger.Write("PsPostingPeriodicTask - ps posting already started, skiping")
		}
	}
}

func GetPsPostGameBundleTask(bundleSize int) taskmanager.SingleTask {
	return func() {
		// var uploadedCovers []string
		var postMessages []string
		gamePostBundle, _ := genegateBundlePostFromSource("ps", bundleSize)

		for _, gamePost := range gamePostBundle {
			msgString := "\n" + gamePost.GameTitle
			msgString += "\nСкидка: " + gamePost.DiscountString + "% \nЦена: " + gamePost.PriceString + " руб."
			msgString += "\nСсылка: " + gamePost.GameURL
			postMessages = append(postMessages, msgString)
			// file.DownloadImage(gamePost.GameCoverURL, "cover_ps_bundle_"+strconv.Itoa(index)+".jpg", func() {
			// 	uploadedCovers = append(uploadedCovers, "cover_ps_bundle_"+strconv.Itoa(index)+".jpg")
			// })
		}

		msgBundleHeader := telegramApi.NewMessageToChannel(CHANNEL_PS_NAME, "Бандл скидок PlayStation Store к этому часу!")
		bot.Send(msgBundleHeader)
		for _, message := range postMessages {
			msgMain := telegramApi.NewMessageToChannel(CHANNEL_PS_NAME, message)
			// msgCover := telegramApi.NewPhotoUpload(CHANNEL_CHAT_ID, uploadedCovers[index])
			// bot.Send(msgCover)
			bot.Send(msgMain)
		}
	}
}