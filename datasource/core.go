package datasource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	filetool "github.com/olegpolukhin/go_ps_scraping/file"
	. "github.com/olegpolukhin/go_ps_scraping/models"
	stringtool "github.com/olegpolukhin/go_ps_scraping/stringtool"
	"github.com/olegpolukhin/go_ps_scraping/taskmanager"

	"github.com/tidwall/gjson"
)

var gamesForPublication []GameGeneral

var activeRepository string

// SetActiveRepository .
func SetActiveRepository(path string) {
	activeRepository = path
}

// GetInitForPublicationTask .
func GetInitForPublicationTask() taskmanager.SingleTask {
	return func() {
		initForPublication()
	}
}

// GetUpdateDiscountedGamesTask .
func GetUpdateDiscountedGamesTask() taskmanager.SingleTask {
	return func() {
		newGames := parseDiscountedGames()
		megreAndSaveGamesToRepository(newGames)
	}
}

func initForPublication() {
	gamesForPublication, _ = loadGamesFromRepo("notYetPublished", 0, 0)
}

// GetRandomDiscountedGame .
func GetRandomDiscountedGame() (game GameGeneral) {
	gamesForPublication, _ = loadGamesFromRepo("notYetPublished", 0, 0)
	if len(gamesForPublication) == 0 {
		return
	}

	i := random(0, len(gamesForPublication))
	game = gamesForPublication[i]
	copy(gamesForPublication[i:], gamesForPublication[i+1:])
	gamesForPublication = gamesForPublication[:len(gamesForPublication)-1]
	updateGameStatusInRepo(game.GlobalID, true)

	return game
}

func updateGameStatusInRepo(globalID string, isPublished bool) {
	gamesFromRepo, isRepoExist := loadGamesFromRepo("all", 0, 0)
	if !isRepoExist {
		return
	}
	gameIndex := GetIndexByGlobalId(gamesFromRepo, globalID)
	if gameIndex >= 0 {
		gamesFromRepo[gameIndex].AlreadyPublished = isPublished
		saveGamesToRepository(gamesFromRepo)
	}
}

func megreAndSaveGamesToRepository(newGames []GameGeneral) {
	oldGames, isRepoFileExist := loadGamesFromRepo("all", 0, 0)
	if !isRepoFileExist {
		saveGamesToRepository(newGames)
		fmt.Println("no games in repo")
	} else {
		fmt.Println("found games in repo")
		mergedGames := MergeGameLists(oldGames, newGames)
		saveGamesToRepository(mergedGames)
	}
}

func parseDiscountedGames() (games []GameGeneral) {
	startPageNumber := 1

	var baseURLSales = viper.GetString("BASE_URL_SALES_PS")
	var baseURLSalesParam = viper.GetString("BASE_URL_SALES_PARAM")

	response, err := http.Get(fmt.Sprintf("%s%s%s", baseURLSales, strconv.Itoa(startPageNumber), baseURLSalesParam))
	if err != nil {
		log.Fatal("response err 1: ", err)
	}
	defer response.Body.Close()

	pageBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	pageBodyString := string(pageBody)
	pagesCount := getDiscountPagesCount(pageBodyString)

	pageBodyString = stringtool.ExtractBetween("grid-cell-container", "grid-footer-controls", pageBodyString)
	gamesRawDataArray := strings.Split(pageBodyString, "class=\"grid-cell grid-cell--game\">")
	gamesRawDataArray = append(gamesRawDataArray[:0], gamesRawDataArray[0+1:]...)
	fmt.Println("ParseDiscountedGames - Found game pages = " + strconv.Itoa(pagesCount))
	loadedGames, idCounter := extractGamesFromRawArray(gamesRawDataArray, 0)
	fmt.Println("ParseDiscountedGames - LoadedGames game pages = " + strconv.Itoa(len(loadedGames)))
	games = append(games, loadedGames...)

	if startPageNumber < pagesCount {
		counter := startPageNumber + 1
		for counter <= pagesCount {
			response, err := http.Get(baseURLSales + strconv.Itoa(counter) + baseURLSalesParam)
			if err != nil {
				log.Fatal("response err 2: ", err)
			}
			defer response.Body.Close()

			// Copy data from the response to standard output
			pageBody, err := ioutil.ReadAll(response.Body)
			defer response.Body.Close()

			pageBodyString := string(pageBody)
			pageBodyString = stringtool.ExtractBetween("grid-cell-container", "grid-footer-controls", pageBodyString)
			gamesRawDataArray := strings.Split(pageBodyString, "class=\"grid-cell grid-cell--game\">")
			gamesRawDataArray = append(gamesRawDataArray[:0], gamesRawDataArray[0+1:]...)
			newGames, nextIDCounter := extractGamesFromRawArray(gamesRawDataArray, idCounter)
			idCounter = nextIDCounter
			games = append(games, newGames...)

			counter++
		}
	} else {
		return games
	}
	return games
}

func extractGamesFromRawArray(gamesRawDataArray []string, startCounter int) (games []GameGeneral, endCounter int) {
	counter := 0
	if startCounter != 0 {
		counter = startCounter
	}
	for _, rawGame := range gamesRawDataArray {
		re := regexp.MustCompile("[0-9]+")

		gameName := stringtool.ExtractBetween("class=\"grid-cell__title \">", "<div class=\"grid-cell__bottom\">", rawGame)
		gameName = strings.Split(gameName, "</span>")[0]
		gameName = strings.Split(gameName, ">")[1]

		gamePriceRaw := stringtool.ExtractBetween("\"price-display__price\">", "button data-fastboot-event-queue=\"add-to-cart", rawGame)
		gamePriceRaw = strings.Split(gamePriceRaw, "</h3>")[0]
		gamePriceRaw = strings.Replace(gamePriceRaw, ".", "", -1)
		if len(re.FindAllString(gamePriceRaw, 1)) >= 1 {
			gamePriceRaw = re.FindAllString(gamePriceRaw, 1)[0]
		} else {
			gamePriceRaw = "0"
		}

		gamePrice, _ := strconv.ParseInt(gamePriceRaw, 10, 14)

		gameDiscountString := stringtool.ExtractBetween("<span class=\"discount-badge__message\">", "<div class=\"grid-cell__body\">", rawGame)
		gameDiscountString = strings.Split(gameDiscountString, "</span>")[0]
		var err error
		var gameDiscount int64

		if strings.Contains(gameDiscountString, "РАСПРОД") {
			gameOldPriceRaw := stringtool.ExtractBetween("<span class=\"price-display__strikethrough\">", "class=\"price-display__price\"", rawGame)
			gameOldPriceRaw = stringtool.ExtractBetween("class=\"price\">", "</div>", gameOldPriceRaw)
			gameOldPriceRaw = strings.Replace(gameOldPriceRaw, ".", "", -1)
			gameOldPriceRaw = re.FindAllString(gameOldPriceRaw, 1)[0]
			oldPrice, parseErr := strconv.ParseInt(gameOldPriceRaw, 10, 14)
			if parseErr != nil {
				gameDiscount = 0
			} else {
				diff := float64(oldPrice - gamePrice)
				gameDiscount = int64((diff / float64(oldPrice)) * 100)
			}
		} else {
			if gameDiscountString == "" {
				gameDiscount = 0
			} else {
				gameDiscountString = re.FindAllString(gameDiscountString, 1)[0]
				gameDiscount, err = strconv.ParseInt(gameDiscountString, 10, 14)
				if err != nil {
					gameDiscount = 0
				}
			}

		}
		gameLinkRaw := strings.Split(rawGame, "grid-cell__prices-container\">")[0]
		gameLinkRaw = strings.Split(rawGame, "class=\"internal-app-link ember-view\">")[0]
		gameLinkRaw = stringtool.ExtractBetween("href=", "id", gameLinkRaw)
		gameLinkRaw = strings.Replace(gameLinkRaw, " ", "", -1)
		gameLinkRaw = strings.Replace(gameLinkRaw, "\"", "", -1)
		gameLink := gameLinkRaw
		gameHeaderRaw := stringtool.ExtractBetween("product-image__img product-image__img--main\">", "<div class=\"product-image__discount-badge\">", rawGame)
		gameHeaderRaw = stringtool.ExtractBetween("3x", "4x", gameHeaderRaw)
		gameHeaderRaw = strings.Replace(gameHeaderRaw, " ", "", -1)
		gameHeaderRaw = strings.Replace(gameHeaderRaw, ",", "", 1)
		gameHeaderRaw = strings.Replace(gameHeaderRaw, "\u0026amp;", "&", -1)
		games = append(games, GameGeneral{
			viper.GetString("GAME_PREFIX") + strconv.Itoa(counter),
			gameName,
			gameHeaderRaw,
			false,
			gameDiscount,
			gamePrice,
			gameLink,
			viper.GetString("GAME_SOURCE"),
			0,
			0,
			false,
			[]int{},
			false,
		})
		counter++
	}
	endCounter = counter
	return games, endCounter
}

func getDiscountPagesCount(rawPageBody string) (count int) {
	if strings.Contains(rawPageBody, "paginator-control__end paginator-control__arrow-navigation") {
		var totalPagesString = strings.Split(rawPageBody, "paginator-control__end paginator-control__arrow-navigation")[0]
		tmpArray := strings.Split(totalPagesString, "grid/STORE-MSF75508-PRICEDROPSCHI")
		totalPagesString = tmpArray[len(tmpArray)-1]
		totalPagesString = strings.Split(totalPagesString, "gameContentType=games")[0]
		logrus.Info("Total pages string = " + totalPagesString)
		re := regexp.MustCompile("[0-9]+")
		totalPagesString = re.FindAllString(totalPagesString, 1)[0]
		var err error
		count, err = strconv.Atoi(totalPagesString)
		if err != nil {
			count = 1
		}
	} else {
		count = 1
	}
	return count
}

func saveGamesToRepository(games []GameGeneral) {
	gamesData := filetool.CreateFile(activeRepository)
	filetool.AppendToFile(gamesData, "{\n\"games\": [\n")
	for index, game := range games {
		gameString, _ := json.Marshal(game)
		gameString = bytes.Replace(gameString, []byte("\\u003c"), []byte("<"), -1)
		gameString = bytes.Replace(gameString, []byte("\\u003e"), []byte(">"), -1)
		gameString = bytes.Replace(gameString, []byte("\\u0026"), []byte("&"), -1)
		filetool.AppendToFile(gamesData, string(gameString))
		if index < (len(games) - 1) {
			filetool.AppendToFile(gamesData, ",\n")
		} else {
			filetool.AppendToFile(gamesData, "\n]\n}")
		}
	}
	filetool.CloseFile(gamesData)
}

func loadGamesFromRepo(loadArgument string, discountBorder int64, discountRange int64) (games []GameGeneral, isRepoFileExist bool) {
	gamesListRaw, loadError := ioutil.ReadFile(activeRepository)
	if loadError != nil {
		return games, isRepoFileExist
	}

	isRepoFileExist = true

	jsonObjects := gjson.Get(string(gamesListRaw), "games")

	for _, object := range jsonObjects.Array() {
		var tmpGame GameGeneral
		json.Unmarshal([]byte(object.String()), &tmpGame)

		switch loadArgument {
		case "all":
			games = append(games, tmpGame)
		case "discounted":
			if tmpGame.Discount <= discountBorder+discountRange && tmpGame.Discount >= discountBorder+discountRange {
				games = append(games, tmpGame)
			}
		case "allDiscounted":
			if tmpGame.Discount > 0 {
				games = append(games, tmpGame)
			}
		case "notYetPublished":
			if !tmpGame.AlreadyPublished {
				games = append(games, tmpGame)
			}
		}
	}

	return games, isRepoFileExist
}

func random(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}
