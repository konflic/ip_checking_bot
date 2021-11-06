package helpers

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
)

type UserRequestEntry struct {
	id         int32
	username   string
	ip_request string
	ip_result  string
	chat_id    string
}

func HasHistory(username string, db *sql.DB) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT id FROM ipbotdb WHERE username = $1);", username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("error checking if row exists %v", err)
	}
	return exists
}

func AlreadyAskedIp(ip_request string, db *sql.DB) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT id FROM ipbotdb WHERE ip_request = $1);", ip_request).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("error checking if row exists %v", err)
	}
	return exists
}

func IsAdmin(username string, db *sql.DB) bool {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT * FROM botadmins WHERE username = $1);", username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("error checking if row exists %v", err)
	}
	return exists
}

func GetIpDataFromDb(ip_request string, db *sql.DB) string {
	var ip_result string
	err := db.QueryRow("SELECT ip_result FROM ipbotdb WHERE ip_request = $1;", ip_request).Scan(&ip_result)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("error getting data from database %v", err)
	}
	return ip_result
}

func AddAdmin(username string, db *sql.DB) {
	_, err := db.Exec("INSERT INTO botadmins (username) VALUES ($1)", username)
	if err != nil {
		log.Fatalf("could not add admin: %v", err)
	}
}

func RemoveAdmin(username string, db *sql.DB) {
	_, err := db.Exec("DELETE FROM botadmins WHERE username = $1;", username)
	if err != nil {
		log.Fatalf("could not insert row: %v", err)
	}
}

func GetDistinctChatIDs(db *sql.DB) (unique_ids []int64) {
	rows, err := db.Query("SELECT DISTINCT chat_id FROM ipbotdb;")

	if err != nil {
		log.Fatalf("could get distinct chat ids: %v", err)
	}

	for rows.Next() {
		var chat_id string
		rows.Scan(&chat_id)
		int_chat_id, _ := strconv.ParseInt(chat_id, 0, 64)
		unique_ids = append(unique_ids, int_chat_id)
	}

	return unique_ids
}

func GetAllUserRequests(username string, db *sql.DB) map[string]string {
	requests := make(map[string]string)
	rows, _ := db.Query("SELECT ip_request, ip_result FROM ipbotdb WHERE username = $1;", username)

	for rows.Next() {
		userEntry := UserRequestEntry{}
		rows.Scan(&userEntry.ip_request, &userEntry.ip_result)
		requests[userEntry.ip_request] = userEntry.ip_result
	}

	return requests
}

func AddRequestEntry(username string, ip_request string, ip_result string, chat_id int64, db *sql.DB) {
	_, err := db.Exec(
		"INSERT INTO ipbotdb (username, ip_request, ip_result, chat_id) VALUES ($1, $2, $3, $4)",
		username, ip_request, ip_result, fmt.Sprint(chat_id))
	if err != nil {
		log.Fatalf("could not insert row: %v", err)
	}
}
