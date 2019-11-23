package taskmanager

import (
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type PeriodicTask func()
type SingleTask func()

func TaskQueue(taskQueue []SingleTask, endMessage string, controlChannel chan string) {
	for _, task := range taskQueue {
		task()
	}
	controlChannel <- endMessage
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
			if time.Now().Hour() == hourToStart {
				logrus.Info("Weekday now " + time.Now().Weekday().String())
				logrus.Info("Time hour now " + strconv.Itoa(time.Now().Hour()))
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
