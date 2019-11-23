package main

import (
	"io"
	"log"
	"os"

	"github.com/olegpolukhin/go_ps_scraping/scheduler"
	"github.com/olegpolukhin/go_ps_scraping/telegram"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("Unable to read config: ", err)
	}
}

func initLogger() {
	file, err := os.OpenFile(viper.GetString("LOG_WARN"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.SetOutput(io.MultiWriter(os.Stdout, file))
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})
}

func main() {
	initConfig()
	initLogger()

	logrus.Info("App started")

	go telegram.BotServerProcess(viper.GetString("BOT_KEY"))

	scheduler.StartScheduler()
}
