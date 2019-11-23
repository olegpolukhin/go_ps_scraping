package scheduler

import (
	"os"

	"github.com/olegpolukhin/go_ps_scraping/datasource"
	"github.com/olegpolukhin/go_ps_scraping/taskmanager"
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

	getUpdateDiscountedTask := datasource.UpdateDiscountedGamesTask()
	getInitPublicationTask := datasource.InitForPublicationTask()
	botPostingBundleTask := telegramBot.SendPostGameBundleTask(viper.GetInt("GAME_BUNDLE"))

	fastStartTasks = append(fastStartTasks,
		getUpdateDiscountedTask,
		getInitPublicationTask,
	)

	postingStartTasks = append(postingStartTasks,
		botPostingBundleTask,
	)

	var playStationUpdateDiscountsTask taskmanager.PeriodicTask = func() {
		taskmanager.TaskQueue(fastStartTasks, "discounts.update", exitChannel)
	}

	var playStationBundlePostingTask taskmanager.PeriodicTask = func() {
		taskmanager.TaskQueue(postingStartTasks, "discounts.update", exitChannel)
	}

	go taskmanager.DoPeriodicTaskAtTime(
		"14:00",
		exitChannel,
		playStationUpdateDiscountsTask,
	)

	go taskmanager.DoPeriodicTaskAtTime(
		"09:00",
		exitChannel,
		playStationBundlePostingTask,
	)

	go taskmanager.DoPeriodicTaskAtTime(
		"10:00",
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
					logrus.Info("All discounts is updated")
				default:
					logrus.Warning(msg)
				}
			}
		}
	}
}
