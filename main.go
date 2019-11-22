package main

import (
	"fmt"
	"log"
	"os"
	"time"

	config "github.com/olegpolukhin/go_ps_scraping/config"
	datasource "github.com/olegpolukhin/go_ps_scraping/datasource"
	taskmanager "github.com/olegpolukhin/go_ps_scraping/taskmanager"
	telegramBot "github.com/olegpolukhin/go_ps_scraping/telegram"

	"github.com/joho/godotenv"
)

func initENV() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config.New()
}

func main() {
	initENV()
	// logger.Init()

	fmt.Println("Search core started. time start: ", time.Now().Format("2006.01.02 15:04:05"))
	exitChannel := make(chan string)

	datasource.PsSetActiveRepository(config.GetEnv.GameDisscounts)

	var psFastStartTasks []taskmanager.SingleTask
	var psPostingStartTasks []taskmanager.SingleTask

	psGameFetchTask := datasource.PsGetUpdateDiscountedGamesTask()
	psInitializationTask := datasource.PsGetInitForPublicationTask()
	botPostingBundlePsTask := telegramBot.GetPsPostGameBundleTask(3)

	telegramBot.BotServerProcess(config.GetEnv.BotKey, exitChannel)

	psFastStartTasks = append(psFastStartTasks,
		psGameFetchTask,
		psInitializationTask,
	)

	psPostingStartTasks = append(psPostingStartTasks,
		botPostingBundlePsTask,
	)

	discountsRefreshEndMessage := "discounts.update"

	var playStationUpdateDiscountsTask taskmanager.PeriodicTask = func() {
		taskmanager.CompleteTaskQueue(psFastStartTasks, discountsRefreshEndMessage, exitChannel)
	}

	var playStationBundlePostingTask taskmanager.PeriodicTask = func() {
		taskmanager.CompleteTaskQueue(psPostingStartTasks, discountsRefreshEndMessage, exitChannel)
	}

	go taskmanager.DoPeriodicTaskAtTime(
		"10:00",
		exitChannel,
		playStationUpdateDiscountsTask,
	)

	go taskmanager.DoPeriodicTaskAtTime(
		"12:00",
		exitChannel,
		playStationBundlePostingTask,
	)

	go taskmanager.DoPeriodicTaskAtTime(
		"14:00",
		exitChannel,
		playStationBundlePostingTask,
	)

	for {
		select {
		case msg := <-exitChannel:
			{
				switch msg {
				case "end":
					fmt.Println("All tasks finished")
					os.Exit(0)
				case "error.network":
					fmt.Println("Network error, mb can`t create request")
				case "discounts.update":
					fmt.Println("All discounts is updated. PlayStation.")
					// logger.Write("All discounts is updated. PlayStation.")
				default:
					// logger.Write(msg)
				}
			}
		}
	}
}
