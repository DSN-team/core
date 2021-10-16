package core

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

//Private variables
var db *sql.DB

//var dbWG *sync.WaitGroup

func createUsersTable() {
	log.Println("creating users table")
	_, err := db.Exec("create table if not exists users (Id integer not null constraint users_pk primary key autoincrement, profile_id integer not null, Username text, Address text not null, public_key text not null, is_friend integer default 0)")
	ErrHandler(err)
	_, err = db.Exec("create unique index if not exists users_id_uindex on users (Id)")
	ErrHandler(err)
}

func createProfilesTable() {
	log.Println("creating Profiles table")
	_, err := db.Exec("create table if not exists Profiles (Id integer not null constraint profiles_pk primary key autoincrement, Username text, Address text, PrivateKey text)")
	ErrHandler(err)
	_, err = db.Exec("create unique index if not exists profiles_id_uindex on Profiles (Id)")
}

func StartDB() {
	log.Println("initing db")
	db, err = sql.Open("sqlite3", "data.db")
	ErrHandler(err)
	_, err := db.Begin()
	ErrHandler(err)
	createProfilesTable()
	createUsersTable()
}

func addUser(user User) {
	log.Println("Adding user", user.Username)
	_, err := db.Exec("INSERT INTO users (profile_id,Username,Address,public_key,is_friend) VALUES ($0,$1,$2,$3,$5)", SelectedProfile.Id, user.Username, user.Address, encPublicKey(marshalPublicKey(user.PublicKey)), user.IsFriend)
	ErrHandler(err)
}

func getFriends() []User {
	log.Println("Getting Friends")
	rows, err := db.Query("SELECT Id, Username, Address, public_key, is_friend FROM users WHERE is_friend = 1 and profile_id = $0", SelectedProfile.Id)
	if ErrHandler(err) {
		return nil
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		ErrHandler(err)
	}(rows)
	var users []User
	for rows.Next() {
		var user User
		var publicKey string
		err = rows.Scan(&user.Id, &user.Username, &user.Address, &publicKey, &user.IsFriend)
		if ErrHandler(err) {
			continue
		}
		decryptedPublicKey := unmarshalPublicKey(decPublicKey(publicKey))
		user.PublicKey = &decryptedPublicKey
		users = append(users, user)
	}
	return users
}

//func getUsers() []User {
//	response, err := db.Query("SELECT * FROM users")
//	if ErrHandler(err) {
//		return nil
//	}
//	var users []User
//	for response.Next() {
//		var user User
//		var PublicKey string
//		err = response.Scan(&user.Username, &user.Address, &PublicKey, &user.IsFriend)
//		if ErrHandler(err) {
//			continue
//		}
//		user.PublicKey = decPublicKey(PublicKey)
//		users = append(users, user)
//	}
//	return users
//}

func addProfile(profile Profile) {
	log.Println("Adding SelectedProfile", profile.Username)
	privateKeyBytes := encProfileKey()
	_, err := db.Exec("INSERT INTO Profiles (Username, Address, PrivateKey) VALUES ($0,$1,$2)", profile.Username, profile.Address, string(privateKeyBytes))
	ErrHandler(err)
}

func getProfiles() []ShowProfile {
	log.Println("Getting Profiles")

	rows, err := db.Query("SELECT Id, Username FROM Profiles")
	if ErrHandler(err) {
		return nil
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		ErrHandler(err)
	}(rows)

	var profiles []ShowProfile
	for rows.Next() {
		var profile ShowProfile
		err = rows.Scan(&profile.Id, &profile.Username)
		if ErrHandler(err) {
			continue
		}
		profiles = append(profiles, profile)
	}
	return profiles
}

func getProfileByID(id int) (string, string, []byte) {
	log.Println("Getting SelectedProfile by ID", id)
	rows, err := db.Query("SELECT Username, Address, PrivateKey FROM Profiles WHERE Id=$0", id)
	if ErrHandler(err) {
		return "", "", nil
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		ErrHandler(err)
	}(rows)

	var username string
	var address string
	var privateKeyString string
	rows.Next()
	err = rows.Scan(&username, &address, &privateKeyString)
	if ErrHandler(err) {
		return "", "", nil
	}

	return username, address, []byte(privateKeyString)
}

func getUserByPublicKey(publicKey string) int {
	log.Println("Getting SelectedProfile by key")
	rows, err := db.Query("SELECT Id FROM users WHERE public_key=$0", publicKey)
	if ErrHandler(err) {
		return 0
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		ErrHandler(err)
	}(rows)

	var id int
	rows.Next()
	err = rows.Scan(&id)
	if ErrHandler(err) {
		return 0
	}
	return id
}
