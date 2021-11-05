package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	ioutil "io/ioutil"
	http "net/http"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const DATABASE_URL = "postgres://postgres:root@localhost:3306/postgres"

type UserRequestEntry struct {
	id         int32
	username   string
	ip_request string
	ip_result  string
	chat_id    string
}

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
		tgbotapi.NewKeyboardButton("Разослать всем"),
		tgbotapi.NewKeyboardButton("Проверить юзера"),
	),
)

func get_ip_data_from_api(ip string) string {
	resp, _ := http.Get("http://ip-api.com/json/" + ip)
	body, _ := ioutil.ReadAll(resp.Body)
	ip_data := string(body)
	return ip_data
}

func get_ip_data_from_db(ip_request string, db *sql.DB) string {
	var ip_result string
	err := db.QueryRow("SELECT ip_result FROM ipbotdb WHERE ip_request = $1;", ip_request).Scan(&ip_result)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("error getting data from database %v", err)
	}
	return ip_result
}

func already_asked_ip(ip_request string, db *sql.DB) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT id FROM ipbotdb WHERE ip_request = $1);", ip_request).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("error checking if row exists %v", err)
	}
	return exists
}

func has_history(username string, db *sql.DB) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT id FROM ipbotdb WHERE username = $1);", username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("error checking if row exists %v", err)
	}
	return exists
}

func is_admin(username string, db *sql.DB) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT * FROM botadmins WHERE username = $1);", username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("error checking if row exists %v", err)
	}
	return exists
}

func get_user_message(updates tgbotapi.UpdatesChannel) string {

	var msg string

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
	bot, err := tgbotapi.NewBotAPI("464078875:AAHhpWqincGZN9J3Wug5dPD5mdkzC6jQFF4")
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
		is_admin := is_admin(username, db)

		switch update.Message.Text {

		case "/start":
			msg := tgbotapi.NewMessage(chat_id, "Привет")
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

			if !already_asked_ip(ip_request, db) {
				ip_result = get_ip_data_from_api(ip_request)
				_, err := db.Exec(
					"INSERT INTO ipbotdb (username, ip_request, ip_result, chat_id) VALUES ($1, $2, $3, $4)",
					username, ip_request, ip_result, fmt.Sprint(chat_id))
				if err != nil {
					log.Fatalf("could not insert row: %v", err)
				}
			} else {
				ip_result = get_ip_data_from_db(ip_request, db)
			}

			bot.Send(tgbotapi.NewMessage(chat_id, ip_result))

		case "Очистить карму":
			msg := tgbotapi.NewMessage(chat_id, "Я тебя не видел...")
			bot.Send(msg)
			_, err := db.Exec("DELETE FROM ipbotdb WHERE username = $1;", username)
			if err != nil {
				log.Fatalf("could not delete row: %v", err)
			}

		case "Вспомнить всё":
			msg := tgbotapi.NewMessage(chat_id, "Подтягиваю базу")
			bot.Send(msg)
			if has_history(username, db) {
				rows, _ := db.Query("SELECT ip_request,  FROM ipbotdb WHERE username = $1;", username)
				for rows.Next() {
					userEntry := UserRequestEntry{}
					rows.Scan(&userEntry.id, &userEntry.username, &userEntry.ip_request, &userEntry.ip_result)
					msg := tgbotapi.NewMessage(chat_id, ">>> "+userEntry.ip_request+" : "+userEntry.ip_result)
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

				_, err := db.Exec("INSERT INTO botadmins (username) VALUES ($1)", add_username)
				if err != nil {
					log.Fatalf("could not insert row: %v", err)
				}
				bot.Send(tgbotapi.NewMessage(chat_id, fmt.Sprintf("Пользователь %s назначен админом!", add_username)))
			}

		case "Удалить админа":
			if is_admin {
				msg := tgbotapi.NewMessage(chat_id, "Кого удалить из админов? (username)")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)

				var remove_username = get_user_message(updates)

				_, err := db.Exec("DELETE FROM botadmins WHERE username = $1;", remove_username)
				if err != nil {
					log.Fatalf("could not insert row: %v", err)
				}

				bot.Send(tgbotapi.NewMessage(chat_id, fmt.Sprintf("Пользователь %s больше не админит.", remove_username)))

			}

		case "Проверить пользователя":
			if is_admin {
				msg := tgbotapi.NewMessage(chat_id, "Какого пользователя проверить? (username)")
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)

				var check_username = get_user_message(updates)

				if has_history(check_username, db) {
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
				rows, _ := db.Query("SELECT DISTINCT chat_id FROM ipbotdb;")

				for rows.Next() {
					var chat_id string
					rows.Scan(&chat_id)
					fmt.Println(chat_id)
					int_chat_id, _ := strconv.ParseInt(chat_id, 0, 64)
					msg := tgbotapi.NewMessage(int_chat_id, "Hello!")
					bot.Send(msg)
				}
			}

		default:
			msg := tgbotapi.NewMessage(chat_id, "Непонятно...")
			bot.Send(msg)
		}
	}
}
