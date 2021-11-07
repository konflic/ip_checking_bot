package helpers

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type UserRequestEntry struct {
	id         int32
	username   string
	ip_request string
	ip_result  string
	chat_id    string
}

const DATABASE_URL = "postgres://postgres:root@127.0.0.1:5432/postgres?sslmode=disable"

func InitDb() *sql.DB {
	db, _ := sql.Open("postgres", DATABASE_URL)

	if err := db.Ping(); err != nil {
		log.Fatalf("%v", err)
	}

	return db
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
		log.Fatal("error checking if row exists %v", err)
	}
	return exists
}

func IsAdmin(username string, db *sql.DB) bool {
	var admin_exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT * FROM botadmins WHERE username = $1);", username).Scan(&admin_exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal("error checking if row exists %v", err)
	}
	return admin_exists
}

func GetIpDataFromDb(ip_request string, db *sql.DB) string {
	var ip_result string
	err := db.QueryRow("SELECT ip_result FROM ipbotdb WHERE ip_request = $1;", ip_request).Scan(&ip_result)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal("error getting data from database %v", err)
	}
	return ip_result
}

func AddAdmin(username string, db *sql.DB) {
	_, err := db.Exec("INSERT INTO botadmins (username) VALUES ($1)", username)
	if err != nil {
		log.Fatal("could not add admin: %v", err)
	}
}

func RemoveAdmin(username string, db *sql.DB) {
	_, err := db.Exec("DELETE FROM botadmins WHERE username = $1;", username)
	if err != nil {
		log.Fatal("could not insert row: %v", err)
	}
}

func GetDistinctChatIDs(db *sql.DB) (unique_ids []int64) {
	rows, err := db.Query("SELECT DISTINCT chat_id FROM ipbotdb;")

	if err != nil {
		log.Fatal("could get distinct chat ids: %v", err)
	}

	for rows.Next() {
		var chat_id string
		rows.Scan(&chat_id)
		int_chat_id, _ := strconv.ParseInt(chat_id, 0, 64)
		unique_ids = append(unique_ids, int_chat_id)
	}

	log.Printf("Got distinc chats: %v", unique_ids)
	return unique_ids
}

func GetDistinctUsernames(db *sql.DB) (unique_usernames []string) {
	rows, err := db.Query("SELECT DISTINCT username FROM ipbotdb;")

	if err != nil {
		log.Fatalf("could get distinct chat ids: %v", err)
	}

	for rows.Next() {
		var username string
		rows.Scan(&username)
		unique_usernames = append(unique_usernames, username)
	}

	return unique_usernames
}

func GetAllUserRequests(username string, db *sql.DB) map[string]string {
	requests := make(map[string]string)
	rows, _ := db.Query("SELECT ip_request, ip_result FROM ipbotdb WHERE username = $1;", username)

	for rows.Next() {
		userEntry := UserRequestEntry{}
		rows.Scan(&userEntry.ip_request, &userEntry.ip_result)
		requests[userEntry.ip_request] = userEntry.ip_result
	}

	log.Printf("Got @%s requests: %v", username, requests)
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

func DeleteUserData(username string, db *sql.DB) {
	_, err := db.Exec("DELETE FROM ipbotdb WHERE username = $1;", username)
	if err != nil {
		log.Fatal("could not clear user requests: %v", err)
	} else {
		log.Printf("All data for user @%s was removed.", username)
	}
}

func DeleteUserIPRequest(username string, ip_request string, db *sql.DB) {
	_, err := db.Exec("DELETE FROM ipbotdb WHERE username = $1 AND ip_request = $2;", username, ip_request)
	if err != nil {
		log.Fatal("could not delete ip request: %v", err)
	} else {
		log.Printf("IP request %s for user @%s was removed.")
	}
}

func SetupDatabase(db *sql.DB) {

	var ipbotdb_exists bool
	var botadmins_exists bool

	if os.Getenv("TELEGRAM_TOKEN") == "" {
		log.Print("You did not set default admin for bot! Set env variable DEFAULT_ADMIN to add admin.")
	}

	db.QueryRow("SELECT EXISTS (SELECT * FROM ipbotdb);").Scan(&ipbotdb_exists)
	db.QueryRow("SELECT EXISTS (SELECT * FROM botadmins);").Scan(&botadmins_exists)

	if !ipbotdb_exists {
		db.Exec("CREATE TABLE ipbotdb (id SERIAL PRIMARY KEY, username VARCHAR(256), ip_request VARCHAR(256), ip_result TEXT, chat_id VARCHAR(256));")
	}

	if !botadmins_exists {
		db.Exec("CREATE TABLE botadmins (id SERIAL PRIMARY KEY, username VARCHAR(256));")
	}

	if !IsAdmin(os.Getenv("DEFAULT_ADMIN"), db) {
		db.Exec("INSERT INTO botadmins (username) VALUES ($1);", os.Getenv("DEFAULT_ADMIN"))
	}

	log.Print("Databases crated and default admin added!")
}
