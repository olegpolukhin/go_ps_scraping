package models

import (
	"strconv"
	"strings"
)

type PostGameDiscount struct {
	HeaderTitle    string
	GameTitle      string
	DiscountString string
	PriceString    string
	GameCoverURL   string
	GameURL        string
}

type GameGeneral struct {
	GlobalID         string      `json:"GlobalID"`
	Name             string      `json:"Name"`
	HeaderImageURL   string      `json:"HeaderImageURL"`
	IsFree           bool        `json:"IsFree"`
	Discount         int64       `json:"Discount"`
	Price            int64       `json:"Price"`
	Link             string      `json:"Link"`
	Source           string      `json:"Source"`
	Metacritic       int64       `json:"Metacritic"`
	SteamID          int64       `json:"SteamID"`
	IsSteamBundle    bool        `json:"IsSteamBundle"`
	SteamBundle      SteamBundle `json:"SteamBundle"`
	AlreadyPublished bool        `json:"AlreadyPublished"`
}

type SteamBundle []int

type Applist struct {
	Applist Apps `json:"applist"`
}

type Apps struct {
	Apps []App `json:"apps"`
}

type App struct {
	AppID int    `json:"appid"`
	Name  string `json:"name"`
}

func MergeGameLists(oldList []GameGeneral, newList []GameGeneral) (mergedList []GameGeneral) {
	mergedList = oldList

	lastGlobalIDString := oldList[(len(oldList) - 1)].GlobalID
	lastGlobalIDString = strings.Split(lastGlobalIDString, "_")[1]

	lastGlobalIDNumber, _ := strconv.Atoi(lastGlobalIDString)
	lastGlobalIDNumber++

	for _, game := range newList {
		if !containsGameGeneral(oldList, game) {
			game.GlobalID = strings.Split(game.GlobalID, "_")[0] + "_" + strconv.Itoa(lastGlobalIDNumber)
			mergedList = append(mergedList, game)
			lastGlobalIDNumber++
		}
	}

	for index, oldGame := range mergedList {
		if !containsGameGeneral(newList, oldGame) {
			copy(mergedList[index:], mergedList[index+1:])
			mergedList = mergedList[:len(mergedList)-1]
		}
	}

	return mergedList
}

func containsGameGeneral(listForCheck []GameGeneral, gameForCheck GameGeneral) bool {
	for _, game := range listForCheck {
		if game.Link == gameForCheck.Link {
			return true
		}
	}
	return false
}

func GetIndexByGlobalId(games []GameGeneral, globalId string) (outIndex int) {
	outIndex = -1
	for index, game := range games {
		if game.GlobalID == globalId {
			return index
		}
	}
	return outIndex
}
