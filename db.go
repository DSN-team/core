package main

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
	_, err := db.Exec("create table if not exists users (id integer not null constraint users_pk primary key autoincrement, username text, address text not null, public_key text not null, is_friend integer default 0)")
	ErrHandler(err)
	_, err = db.Exec("create unique index if not exists users_id_uindex on users (id)")
	ErrHandler(err)
}

func createProfilesTable() {
	log.Println("creating profiles table")
	_, err := db.Exec("create table if not exists profiles (id integer not null constraint profiles_pk primary key autoincrement, username text, address text, privateKey text)")
	ErrHandler(err)
	_, err = db.Exec("create unique index if not exists profiles_id_uindex on profiles (id)")
}

func startDB() {
	log.Println("initing db")
	db, err = sql.Open("sqlite3", "data.db")
	ErrHandler(err)
	_, err := db.Begin()
	ErrHandler(err)
	createProfilesTable()
	createUsersTable()
}

func addUser(user User) {
	log.Println("Adding user", user.username)
	_, err := db.Exec("INSERT INTO users (username,address,public_key,is_friend) VALUES ($0,$1,$2,$3)", user.username, user.address, encPublicKey(marshalPublicKey(user.publicKey)), user.isFriend)
	ErrHandler(err)
}

func getFriends() []User {
	log.Println("Getting friends")
	rows, err := db.Query("SELECT id, username, address, public_key, is_friend FROM users WHERE is_friend = 1")
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
		err = rows.Scan(&user.id, &user.username, &user.address, &publicKey, &user.isFriend)
		if ErrHandler(err) {
			continue
		}
		decryptedPublicKey := unmarshalPublicKey(decPublicKey(publicKey))
		user.publicKey = &decryptedPublicKey
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
//		var publicKey string
//		err = response.Scan(&user.username, &user.address, &publicKey, &user.isFriend)
//		if ErrHandler(err) {
//			continue
//		}
//		user.publicKey = decPublicKey(publicKey)
//		users = append(users, user)
//	}
//	return users
//}

func addProfile(profile Profile) {
	log.Println("Adding profile", profile.username)
	privateKeyBytes := encProfileKey()
	_, err := db.Exec("INSERT INTO profiles (username, address, privateKey) VALUES ($0,$1,$2)", profile.username, profile.address, string(privateKeyBytes))
	ErrHandler(err)
}

func getProfiles() []ShowProfile {
	log.Println("Getting profiles")

	rows, err := db.Query("SELECT id, username FROM profiles")
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
		err = rows.Scan(&profile.id, &profile.username)
		if ErrHandler(err) {
			continue
		}
		profiles = append(profiles, profile)
	}
	return profiles
}

func getProfileByID(id int) (string, string, []byte) {
	log.Println("Getting profile by ID", id)
	rows, err := db.Query("SELECT username, address, privateKey FROM profiles WHERE id=$0", id)
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
	log.Println("Getting profile by key")
	rows, err := db.Query("SELECT id FROM users WHERE public_key=$0", publicKey)
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
