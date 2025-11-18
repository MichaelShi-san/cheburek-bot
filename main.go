package main

import (
	"log"
	"os"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type State string

const (
	StateStart       State = "start"
	StateAreYouCheb  State = "are_you_cheb"
	StateHalal       State = "halal"
	StateMeatChoice  State = "meat_choice"
)

var (
	userState = make(map[int64]State)
	mu        sync.Mutex
)

func setState(chatID int64, state State) {
	mu.Lock()
	defer mu.Unlock()
	userState[chatID] = state
}

func getState(chatID int64) State {
	mu.Lock()
	defer mu.Unlock()
	return userState[chatID]
}

func main() {
	// 1. Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env not found, using system variables")
	}

	// 2. Получаем токен
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("❌ BOT_TOKEN is empty! Check your .env")
	}

	// 3. Создаём бота
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("🤖 Bot started:", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		// /start
		if update.Message != nil && update.Message.Text == "/start" {
			askAreYouCheb(bot, update.Message.Chat.ID)
			continue
		}

		// inline buttons
		if update.CallbackQuery != nil {
			chatID := update.CallbackQuery.Message.Chat.ID
			data := update.CallbackQuery.Data
			state := getState(chatID)

			switch state {

			case StateAreYouCheb:
				if data == "no" {
					bot.Send(tgbotapi.NewMessage(chatID, "Ты не чебурек ❌"))
					setState(chatID, StateStart)
				}
				if data == "yes" {
					askHalal(bot, chatID)
				}

			case StateHalal:
				if data == "not_halal" {
					bot.Send(tgbotapi.NewMessage(chatID, "Вы — чебурек со свининой 🐷"))
					setState(chatID, StateStart)
				}
				if data == "halal" {
					askMeat(bot, chatID)
				}

			case StateMeatChoice:
				if data == "chicken" {
					bot.Send(tgbotapi.NewMessage(chatID, "Вы — чебурек с курицей 🐔"))
					setState(chatID, StateStart)
				}
				if data == "beef" {
					bot.Send(tgbotapi.NewMessage(chatID, "Вы — чебурек с говядиной 🐄"))
					setState(chatID, StateStart)
				}
			}
		}
	}
}

func askAreYouCheb(bot *tgbotapi.BotAPI, chatID int64) {
	setState(chatID, StateAreYouCheb)
	msg := tgbotapi.NewMessage(chatID, "Ты чебурек?")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", "yes"),
			tgbotapi.NewInlineKeyboardButtonData("Нет", "no"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func askHalal(bot *tgbotapi.BotAPI, chatID int64) {
	setState(chatID, StateHalal)
	msg := tgbotapi.NewMessage(chatID, "Какой ты чебурек?")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Халяль", "halal"),
			tgbotapi.NewInlineKeyboardButtonData("Не халяль", "not_halal"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func askMeat(bot *tgbotapi.BotAPI, chatID int64) {
	setState(chatID, StateMeatChoice)
	msg := tgbotapi.NewMessage(chatID, "Выберите начинку:")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Курица", "chicken"),
			tgbotapi.NewInlineKeyboardButtonData("Говядина", "beef"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}
