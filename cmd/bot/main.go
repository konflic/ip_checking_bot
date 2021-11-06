package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/konflic/ip_checking_bot/helpers"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const DATABASE_URL = "postgres://postgres:root@localhost:3306/postgres"

var numericKeyboardUser = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Пробить IP"),
		tgbotapi.NewKeyboardButton("Вспомнить всё"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Очистить карму"),
	),
)

var numericKeyboardAdmin = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Пробить IP"),
		tgbotapi.NewKeyboardButton("Вспомнить всё"),
		tgbotapi.NewKeyboardButton("Очистить карму"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Добавить админа"),
		tgbotapi.NewKeyboardButton("Удалить админа"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Рассылка"),
		tgbotapi.NewKeyboardButton("Проверить юзера"),
	),
)

func get_ip_data_from_api(ip string) string {
	resp, _ := http.Get("http://ip-api.com/json/" + ip)
	body, _ := ioutil.ReadAll(resp.Body)
	ip_data := string(body)
	return ip_data
}

func get_user_message(updates tgbotapi.UpdatesChannel) (msg string) {

	for update := range updates {
		if update.Message == nil {
			// ignore any non-Message Updates
			continue
		}
		msg = update.Message.Text
		break
	}

	return msg
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	db, _ := sql.Open("pgx", DATABASE_URL)

	// Setting debug mod
	bot.Debug = false

	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("%v", err)
	}

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil {
			// ignore any non-Message Updates
			continue
		}

		chat_id := update.Message.Chat.ID
		username := update.Message.From.UserName
		is_admin := helpers.IsAdmin(username, db)

		switch update.Message.Text {

		case "/start":
			msg := tgbotapi.NewMessage(chat_id, "Привет, я бот для вычисления по IP.")
			msg.ReplyToMessageID = update.Message.MessageID
			if is_admin {
				msg.ReplyMarkup = numericKeyboardAdmin
			} else {
				msg.ReplyMarkup = numericKeyboardUser
			}
			bot.Send(msg)

		case "Пробить IP":
			msg := tgbotapi.NewMessage(chat_id, "Какой IP пробить? (X.X.X.X)")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)

			var ip_result string
			var ip_request = get_user_message(updates)

			if net.ParseIP(ip_request) != nil {
				if !helpers.AlreadyAskedIp(ip_request, db) {
					ip_result = get_ip_data_from_api(ip_request)
					helpers.AddRequestEntry(username, ip_request, ip_result, chat_id, db)
				} else {
					// Если такой ip уже пробивали то подтягиваем из базы
					// TODO: Запоминать время последнего пробива, чтобы обновлять кэш
					ip_result = helpers.GetIpDataFromDb(ip_request, db)
				}
				bot.Send(tgbotapi.NewMessage(chat_id, ip_result))
			} else {
				bot.Send(tgbotapi.NewMessage(chat_id, "Это не ip адресс"))
			}

		case "Очистить карму":
			_, err := db.Exec("DELETE FROM ipbotdb WHERE username = $1;", username)
			if err != nil {
				log.Fatalf("could not clear user requests: %v", err)
			}
			msg := tgbotapi.NewMessage(chat_id, "Я тебя не видел...")
			bot.Send(msg)

		case "Вспомнить всё":
			msg := tgbotapi.NewMessage(chat_id, "Подтягиваю базу")
			bot.Send(msg)
			if helpers.HasHistory(username, db) {
				for ip_request, ip_result := range helpers.GetAllUserRequests(username, db) {
					msg := tgbotapi.NewMessage(chat_id, ">>> "+ip_request+" : "+ip_result)
					bot.Send(msg)
				}
			} else {
				bot.Send(tgbotapi.NewMessage(chat_id, "ничего..."))
			}

		case "Добавить админа":
			if is_admin {
				msg := tgbotapi.NewMessage(chat_id, "Кого добавить в админы? (username)")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)

				var add_username = get_user_message(updates)
				helpers.AddAdmin(add_username, db)

				bot.Send(tgbotapi.NewMessage(chat_id, fmt.Sprintf("Пользователь %s назначен админом!", add_username)))
			}

		case "Удалить админа":
			if is_admin {
				msg := tgbotapi.NewMessage(chat_id, "Кого удалить из админов? (username)")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)

				var remove_username = get_user_message(updates)
				helpers.RemoveAdmin(remove_username, db)

				bot.Send(tgbotapi.NewMessage(chat_id, fmt.Sprintf("Пользователь %s больше не админит.", remove_username)))

			}

		case "Проверить юзера":
			if is_admin {
				msg := tgbotapi.NewMessage(chat_id, "Какого пользователя проверить? (username)")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)

				var check_username = get_user_message(updates)

				if helpers.HasHistory(check_username, db) {
					rows, _ := db.Query("SELECT ip_request FROM ipbotdb WHERE username = $1;", username)
					for rows.Next() {
						var ip_request string
						rows.Scan(&ip_request)
						msg := tgbotapi.NewMessage(chat_id, ">>> "+ip_request)
						bot.Send(msg)
					}
				} else {
					bot.Send(tgbotapi.NewMessage(chat_id, "на него ничего нет..."))
				}
			}

		case "Разослать всем":
			if is_admin {
				msg := tgbotapi.NewMessage(chat_id, "Какое сообщение разослать?")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)

				message_to_all := get_user_message(updates)

				for chat_id := range helpers.GetDistinctChatIDs(db) {
					msg := tgbotapi.NewMessage(int64(chat_id), message_to_all)
					bot.Send(msg)
				}
			}

		default:
			msg := tgbotapi.NewMessage(chat_id, "Непонятно...")
			bot.Send(msg)
		}
	}
}
