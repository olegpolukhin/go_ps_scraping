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

	"github.com/olegpolukhin/go_ps_scraping/config"
	filetool "github.com/olegpolukhin/go_ps_scraping/file"
	. "github.com/olegpolukhin/go_ps_scraping/models"
	stringtool "github.com/olegpolukhin/go_ps_scraping/stringtool"
	"github.com/olegpolukhin/go_ps_scraping/taskmanager"

	"github.com/tidwall/gjson"
)

var gamesForPublicationPs []GameGeneral

var activeRepositoryPs string

func PsSetActiveRepository(path string) {
	activeRepositoryPs = path
}

func PsGetInitForPublicationTask() taskmanager.SingleTask {
	return func() {
		PsInitForPublication()
	}
}

func PsGetUpdateDiscountedGamesTask() taskmanager.SingleTask {
	newGames := PsParseDiscountedGames()
	if newGames != nil {
		PsMegreAndSaveGamesToRepository(newGames)
	}
	return func() {
		newGames := PsParseDiscountedGames()
		PsMegreAndSaveGamesToRepository(newGames)
	}
}

func PsInitForPublication() {
	gamesForPublicationPs, _ = PsLoadGamesFromRepo("notYetPublished", 0, 0)
	// logger.Write("PsInitForPublication.Inited - " + strconv.Itoa(len(gamesForPublicationPs)) + " games ready.")
}

func PsGetRandomDiscountedGame() (game GameGeneral) {
	gamesForPublicationPs, _ = PsLoadGamesFromRepo("notYetPublished", 0, 0)
	if len(gamesForPublicationPs) == 0 {
		return
	}

	i := randomPs(0, len(gamesForPublicationPs))
	game = gamesForPublicationPs[i]
	copy(gamesForPublicationPs[i:], gamesForPublicationPs[i+1:])
	gamesForPublicationPs = gamesForPublicationPs[:len(gamesForPublicationPs)-1]
	PsUpdateGameStatusInRepo(game.GlobalID, true)

	return game
}

func PsUpdateGameStatusInRepo(globalId string, isAlreadyPublished bool) {
	gamesFromRepo, isRepoExist := PsLoadGamesFromRepo("all", 0, 0)
	if !isRepoExist {
		return
	}
	gameIndex := GetIndexByGlobalId(gamesFromRepo, globalId)
	if gameIndex >= 0 {
		gamesFromRepo[gameIndex].AlreadyPublished = isAlreadyPublished
		PsSaveGamesToRepository(gamesFromRepo)
	}
}

func PsMegreAndSaveGamesToRepository(newGames []GameGeneral) {
	oldGames, isRepoFileExist := PsLoadGamesFromRepo("all", 0, 0)
	if !isRepoFileExist {
		PsSaveGamesToRepository(newGames)
		fmt.Println("no games in repo")
	} else {
		fmt.Println("found games in repo")
		mergedGames := MergeGameLists(oldGames, newGames)
		PsSaveGamesToRepository(mergedGames)
	}
}

func PsParseDiscountedGames() (games []GameGeneral) {
	startPageNumber := 1

	var baseURLSales = config.GetEnv.BaseURLSales
	var baseURLSalesParam = config.GetEnv.BaseURLSalesParam

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
	pagesCount := psGetDiscountPagesCount(pageBodyString)

	pageBodyString = stringtool.ExtractBetween("grid-cell-container", "grid-footer-controls", pageBodyString)
	gamesRawDataArray := strings.Split(pageBodyString, "class=\"grid-cell grid-cell--game\">")
	gamesRawDataArray = append(gamesRawDataArray[:0], gamesRawDataArray[0+1:]...)
	fmt.Println("PlayStationCore.PsParseDiscountedGames - Found game pages = " + strconv.Itoa(pagesCount))
	loadedGames, idCounter := psExtractGamesFromRawArray(gamesRawDataArray, 0)
	fmt.Println("PlayStationCore.PsParseDiscountedGames - Found game pages = " + strconv.Itoa(len(loadedGames)))
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
			if err != nil {
				// нет обработки ошибок
			}
			defer response.Body.Close()

			pageBodyString := string(pageBody)
			pageBodyString = stringtool.ExtractBetween("grid-cell-container", "grid-footer-controls", pageBodyString)
			gamesRawDataArray := strings.Split(pageBodyString, "class=\"grid-cell grid-cell--game\">")
			gamesRawDataArray = append(gamesRawDataArray[:0], gamesRawDataArray[0+1:]...)
			newGames, nextIDCounter := psExtractGamesFromRawArray(gamesRawDataArray, idCounter)
			idCounter = nextIDCounter
			games = append(games, newGames...)

			counter++
		}
	} else {
		return games
	}
	return games
}

func psExtractGamesFromRawArray(gamesRawDataArray []string, startCounter int) (games []GameGeneral, endCounter int) {
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
		} else { // этот else можно сократить
			if gameDiscountString == "" {
				gameDiscount = 0
			} else { // и этот
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
			config.GetEnv.GameIDPrefix + strconv.Itoa(counter),
			gameName,
			gameHeaderRaw,
			false,
			gameDiscount,
			gamePrice,
			gameLink,
			config.GetEnv.GameSource,
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

func psGetDiscountPagesCount(rawPageBody string) (count int) {
	if strings.Contains(rawPageBody, "paginator-control__end paginator-control__arrow-navigation") {
		var totalPagesString = strings.Split(rawPageBody, "paginator-control__end paginator-control__arrow-navigation")[0]
		tmpArray := strings.Split(totalPagesString, "grid/STORE-MSF75508-PRICEDROPSCHI")
		totalPagesString = tmpArray[len(tmpArray)-1]
		totalPagesString = strings.Split(totalPagesString, "gameContentType=games")[0]
		fmt.Println("Total PS pages string = " + totalPagesString)
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

func PsSaveGamesToRepository(games []GameGeneral) {
	gamesData := filetool.CreateFile(activeRepositoryPs)
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

func PsLoadGamesFromRepo(loadArgument string, discountBorder int64, discountRange int64) (games []GameGeneral, isRepoFileExist bool) {
	isRepoFileExist = false
	gamesListRaw, loadError := ioutil.ReadFile(activeRepositoryPs)
	if loadError != nil {
		return games, isRepoFileExist
	} else {
		isRepoFileExist = true
	}

	jsonObjects := gjson.Get(string(gamesListRaw), "games")

	for _, object := range jsonObjects.Array() {
		var tmpGame GameGeneral
		err := json.Unmarshal([]byte(object.String()), &tmpGame)
		if err != nil {
			// нет обработки
		}

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

func randomPs(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}
