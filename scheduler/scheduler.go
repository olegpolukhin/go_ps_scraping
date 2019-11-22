package scheduler

import (
	"os"

	datasource "github.com/olegpolukhin/go_ps_scraping/datasource"
	taskmanager "github.com/olegpolukhin/go_ps_scraping/taskmanager"
	telegramBot "github.com/olegpolukhin/go_ps_scraping/telegram"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// StartScheduler Procedure start scheduler. Performs scheduled tasks
func StartScheduler() {
	exitChannel := make(chan string)
	datasource.SetActiveRepository(viper.GetString("GAME_DISSCOUNTS"))
	var fastStartTasks []taskmanager.SingleTask
	var postingStartTasks []taskmanager.SingleTask

	psGameFetchTask := datasource.GetUpdateDiscountedGamesTask()
	psInitializationTask := datasource.GetInitForPublicationTask()
	botPostingBundleTask := telegramBot.SendPostGameBundleTask(3)

	fastStartTasks = append(fastStartTasks,
		psGameFetchTask,
		psInitializationTask,
	)

	postingStartTasks = append(postingStartTasks,
		botPostingBundleTask,
	)

	var playStationUpdateDiscountsTask taskmanager.PeriodicTask = func() {
		taskmanager.CompleteTaskQueue(fastStartTasks, "discounts.update", exitChannel)
	}

	var playStationBundlePostingTask taskmanager.PeriodicTask = func() {
		telegramBot.BotServerProcess(viper.GetString("BOT_KEY"), exitChannel)
		taskmanager.CompleteTaskQueue(postingStartTasks, "discounts.update", exitChannel)
	}

	go taskmanager.DoPeriodicTaskAtTime(
		"09:00",
		exitChannel,
		playStationUpdateDiscountsTask,
	)

	go taskmanager.DoPeriodicTaskAtTime(
		"12:00",
		exitChannel,
		playStationBundlePostingTask,
	)

	go taskmanager.DoPeriodicTaskAtTime(
		"17:00",
		exitChannel,
		playStationBundlePostingTask,
	)

	for {
		select {
		case msg := <-exitChannel:
			{
				switch msg {
				case "end":
					logrus.Info("All tasks finished")
					os.Exit(0)
				case "error.network":
					logrus.Error("Network error: Create request")
				case "discounts.update":
					datasource.GetUpdateDiscountedGamesTask()
					datasource.GetInitForPublicationTask()
					logrus.Info("All discounts is updated")
				default:
					logrus.Warning(msg)
				}
			}
		}
	}
}
