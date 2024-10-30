package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"myxxa/scheduler/internal/db"
	"myxxa/scheduler/internal/network"
	"myxxa/scheduler/internal/schedule"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewInstance() (*tgbotapi.BotAPI, error) {
	tgApi, exists := os.LookupEnv("tgApi")
	if !exists {
		log.Fatal(errors.New(".env не нашел tgApi"))
	}
	bot, err := tgbotapi.NewBotAPI(tgApi)
	if err != nil {
		return nil, err
	}
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return bot, nil
}

func StartBot(ctx context.Context, bot *tgbotapi.BotAPI, cache *db.Cache) {
	var list = []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "Начало работы бота",
		},
		{
			Command:     "help",
			Description: "Показать справочное сообщение",
		},
	}
	bot.Request(tgbotapi.NewSetMyCommands(
		list...,
	))
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	defer bot.StopReceivingUpdates()

	for update := range updates {
		go func() {
			network.InterceptorSchedule(cache)
			handleUpdate(ctx, bot, &update, cache)
		}()
	}
}

var currentWeekTypeSchedule = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Понедельник"),
		tgbotapi.NewKeyboardButton("Вторник"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Среда"),
		tgbotapi.NewKeyboardButton("Четверг"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Пятница"),
		tgbotapi.NewKeyboardButton("Следующая неделя"),
	),
)

var nextWeekTypeSchedule = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Понедельник*"),
		tgbotapi.NewKeyboardButton("Вторник*"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Среда*"),
		tgbotapi.NewKeyboardButton("Четверг*"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Пятница*"),
		tgbotapi.NewKeyboardButton("Текущая неделя"),
	),
)

func handleUpdate(_ context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update, cache *db.Cache) {
	if update.Message != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ParseMode = "markdown"
		currentWeek := schedule.GetWeekType()
		var nextWeek schedule.WeekType
		switch currentWeek {
		case schedule.WeekTypeNumerator:
			nextWeek = schedule.WeekTypeDenominator
			break
		case schedule.WeekTypeDenominator:
			nextWeek = schedule.WeekTypeNumerator
			break
		default:
			break
		}
		switch update.Message.Text {
		case "/start":
			startHandle(currentWeek, &msg)
			break
		case "/help":
			response, found := cache.GetUnmarshal("schedule 0")
			if !found {
				log.Fatal(errors.New("Ключ не найден").Error())
			}
			schedule.MapToSchedule(response)
			break
		case "Понедельник":
			text := db.GetDaySchedule(0, currentWeek, cache)
			msg.Text = fmt.Sprintf("Расписание на _Понедельник_ %v", text)
			break
		case "Вторник":
			text := db.GetDaySchedule(1, currentWeek, cache)
			msg.Text = fmt.Sprintf("Расписание на _Вторник_ %v", text)
			break
		case "Среда":
			text := db.GetDaySchedule(2, currentWeek, cache)
			msg.Text = fmt.Sprintf("Расписание на _Среда_ %v", text)
			break
		case "Четверг":
			text := db.GetDaySchedule(3, currentWeek, cache)
			msg.Text = fmt.Sprintf("Расписание на _Четверг_ %v", text)
			break
		case "Пятница":
			text := db.GetDaySchedule(4, currentWeek, cache)
			msg.Text = fmt.Sprintf("Расписание на _Пятница_ %v", text)
			break
		case "Понедельник*":
			text := db.GetDaySchedule(0, nextWeek, cache)
			msg.Text = fmt.Sprintf("Расписание на _Понедельник_ %v", text)
			break
		case "Вторник*":
			text := db.GetDaySchedule(1, nextWeek, cache)
			msg.Text = fmt.Sprintf("Расписание на _Вторник_ %v", text)
			break
		case "Среда*":
			text := db.GetDaySchedule(2, nextWeek, cache)
			msg.Text = fmt.Sprintf("Расписание на _Среда_ %v", text)
			break
		case "Четверг*":
			text := db.GetDaySchedule(3, nextWeek, cache)
			msg.Text = fmt.Sprintf("Расписание на _Четверг_ %v", text)
			break
		case "Пятница*":
			text := db.GetDaySchedule(4, nextWeek, cache)
			msg.Text = fmt.Sprintf("Расписание на _Пятница_ %v", text)
			break
		case "Следующая неделя":
			startHandle(nextWeek, &msg)
			break
		case "Текущая неделя":
			startHandle(currentWeek, &msg)
			break
		}
		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}
	}
}

func startHandle(week schedule.WeekType, msg *tgbotapi.MessageConfig) {
	dt := time.Now().Format("_02.01.2006 15:04_\n")

	var weekStringType string
	switch week {
	case (schedule.WeekTypeNumerator):
		weekStringType = "Числитель"
	case (schedule.WeekTypeDenominator):
		weekStringType = "Знаменатель"
	}
	currentWeek := schedule.GetWeekType()
	if currentWeek == week {
		msg.ReplyMarkup = currentWeekTypeSchedule
		now := time.Now()
		start := getStartOfWeek(now)
		end := getEndOfWeek(start)
		dateWeek := fmt.Sprintf("\nНеделя: %s - %s", start.Format("02.01.2006"), end.Format("02.01.2006"))
		msg.Text = dt + fmt.Sprintf("\nТекущая неделя: *%v*", weekStringType) + dateWeek + "\nВыберите день:"
	} else {
		msg.ReplyMarkup = nextWeekTypeSchedule
		now := time.Now().AddDate(0, 0, 7)
		start := getStartOfWeek(now)
		end := getEndOfWeek(start)
		dateWeek := fmt.Sprintf("\nНеделя: %s - %s", start.Format("02.01.2006"), end.Format("02.01.2006"))
		msg.Text = dt + fmt.Sprintf("\nСледующая неделя: *%v*", weekStringType) + dateWeek + "\nВыберите день:"
	}
}

// Функция для нахождения начала недели (понедельник)
func getStartOfWeek(date time.Time) time.Time {
	for date.Weekday() != time.Monday { // Сдвигаемся назад до первого дня недели (понедельника)
		date = date.AddDate(0, 0, -1)
	}
	return date
}

// Функция для нахождения конца недели (воскресенье)
func getEndOfWeek(start time.Time) time.Time {
	end := start.AddDate(0, 0, 6) // Добавляем 6 дней к началу недели
	return end
}
