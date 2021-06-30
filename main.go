package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type wallet map[string]float64

var db = map[int64]wallet{}

type bnResponse struct {
	Price float64 `json:"price,string"`
	Code  int64   `json:"code"`
}

func main() {
	bot, err := tgbotapi.NewBotAPI("1826562478:AAH1tCY41Nij5IpQ3COzNzGf0DCzbMNHd6o")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		command := strings.Split(update.Message.Text, " ")

		switch command[0] {
		case "ADD":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "wrong command"))
			}
			amount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}
			if _, ok := db[update.Message.Chat.ID]; !ok {
				db[update.Message.Chat.ID] = wallet{}
			}
			db[update.Message.Chat.ID][command[1]] += amount
			balanceText := fmt.Sprintf("%f", db[update.Message.Chat.ID][command[1]])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, balanceText))

		case "SUB":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "wrong command"))
			}
			amount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}
			if _, ok := db[update.Message.Chat.ID]; !ok {
				db[update.Message.Chat.ID] = wallet{}
			}
			db[update.Message.Chat.ID][command[1]] -= amount
			if amount < 0 {
				amount = 0
			}
			balanceText := fmt.Sprintf("%f", db[update.Message.Chat.ID][command[1]])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, balanceText))
		case "DEL":
			if len(command) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "wrong command"))
			}
			delete(db[update.Message.Chat.ID], command[1])
			//bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "command not found"))
		case "SHOW":
			msg := ""
			var sum float64 = 0
			for key, value := range db[update.Message.Chat.ID] {
				price, err := getPrice(key)
				log.Printf("[%s] %f", err, price)
				sum += value * price
				msg += fmt.Sprintf("%s: %f [%.2f]\n", key, value, value*price)
			}
			msg += fmt.Sprintf("Total: %.2f \n", sum)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "command not found"))
		}
		log.Printf("[%s] %s", update.Message.From.UserName, command[0])

		//msg := tgbotapi.NewMessage(update.Message.Chat.ID, command[0])
		//msg.ReplyToMessageID = update.Message.MessageID

		//bot.Send(msg)
	}
}

func getPrice(symbol string) (price float64, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol))
	log.Printf("[%s]", symbol)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var jsonResp bnResponse

	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}

	if jsonResp.Code != 0 {
		err = errors.New("wrong symbol")
	}

	price = jsonResp.Price

	return
}
