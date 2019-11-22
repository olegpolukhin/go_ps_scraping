package taskmanager

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var TIME_DIFF_MSK = 0 // 3
var TIME_DIFF_EKB = 5

type PeriodicTask func()
type SingleTask func()

type ChainTask struct {
	StartTask SingleTask
	NextTask  SingleTask
}

func CompleteTaskQueue(taskQueue []SingleTask, endMessage string, controlChannel chan string) {
	for _, task := range taskQueue {
		task()
	}
	controlChannel <- endMessage
}

//Need to fix this shit
func StartPeriodicTask(taskExecutionPeriod int64, timePeriodType string, workingPeriodStartHour int, workingPeriodEndHour int, controlChannel chan string, task PeriodicTask) {
	var ticker *time.Ticker
	// не ясно зачем стак сделанно если в if все гда true то его можно убрать
	// var firstStart bool = true

	switch timePeriodType {
	case "second":
		ticker = time.NewTicker(time.Duration(taskExecutionPeriod) * time.Second)
	case "hour":
		ticker = time.NewTicker(time.Duration(taskExecutionPeriod) * time.Hour)
	default:
		return
	}
	// и логика не изменилась
	// if firstStart {
	var timeHourNow = time.Now().Hour()
	fmt.Println("Date now " + time.Now().Weekday().String())
	fmt.Println("Time hour now " + strconv.Itoa(timeHourNow+TIME_DIFF_MSK))
	if timeHourNow+TIME_DIFF_MSK > workingPeriodStartHour && timeHourNow+TIME_DIFF_MSK < workingPeriodEndHour {
		task()
		fmt.Println("Task started")
		// не понятно что за присвоение и для чего
		// эта переменная не используеться больше
		// firstStart = false
	} else {
		fmt.Println("Task out of working period, waiting for next tick")
	}
	// }

	for {
		select {
		case <-ticker.C:
			var timeHourNow = time.Now().Hour()
			fmt.Println("Date now " + time.Now().Weekday().String())
			fmt.Println("Time hour now " + strconv.Itoa(timeHourNow+TIME_DIFF_MSK))
			if timeHourNow+TIME_DIFF_MSK > workingPeriodStartHour && timeHourNow+TIME_DIFF_MSK < workingPeriodEndHour {
				task()
			}
		case <-controlChannel:
			msg := <-controlChannel
			if msg == "stop" {
				ticker.Stop()
				return
			}
		}
	}
}

func DoPeriodicTaskAtTime(timeToStart string, controlChannel chan string, task PeriodicTask) {
	var ticker = time.NewTicker(time.Duration(1) * time.Minute)
	var hourToStart int
	var minuteToStart int
	var hourParseErr, minuteParseErr error
	timeToStartRaw := strings.Split(timeToStart, ":")

	hourToStart, hourParseErr = strconv.Atoi(timeToStartRaw[0])
	if hourParseErr != nil {
		return
	} else if hourToStart < 0 || hourToStart > 23 {
		return
	}
	minuteToStart, minuteParseErr = strconv.Atoi(timeToStartRaw[1])
	if minuteParseErr != nil {
		return
	} else if minuteToStart < 0 || minuteToStart > 59 {
		return
	}

	for {
		select {
		case <-ticker.C:
			var timeHourNow = time.Now().Hour()
			var timeMinuteNow = time.Now().Minute()
			fmt.Println("Weekday now " + time.Now().Weekday().String())
			fmt.Println("Time hour now " + strconv.Itoa(timeHourNow))
			if timeHourNow == hourToStart && timeMinuteNow == minuteToStart {
				task()
			}
		case <-controlChannel:
			msg := <-controlChannel
			if msg == "stop" {
				ticker.Stop()
				return
			}
		}
	}
}
