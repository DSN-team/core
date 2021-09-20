package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

//Private variables
var db *sql.DB

func createUsersTable() {
	_, err := db.Exec("create table if not exists users (id integer not null constraint users_pk primary key autoincrement, username text, address text not null, public_key text not null, is_friend integer default 0)")
	ErrHandler(err)
	_, err = db.Exec("create unique index if not exists users_id_uindex on users (id)")
	ErrHandler(err)
}

func createProfilesTable() {
	_, err := db.Exec("create table if not exists profiles (id integer not null constraint profiles_pk primary key autoincrement, username text, address text, privateKey text)")
	ErrHandler(err)
	_, err = db.Exec("create unique index if not exists profiles_id_uindex on profiles (id)")
}

func startDB() {
	db, err = sql.Open("sqlite3", "data.db")
	ErrHandler(err)
	_, err := db.Begin()
	ErrHandler(err)
	createProfilesTable()
	createUsersTable()
}

func addUser(user User) {
	_, err := db.Exec("INSERT INTO users (username,address,public_key,is_friend) VALUES ($0,$1,$2,$3)", user.username, user.address, encPublicKey(*user.publicKey), user.isFriend)
	ErrHandler(err)
}

func getFriends() []User {
	response, err := db.Query("SELECT * FROM users WHERE is_friend = true")
	if ErrHandler(err) {
		return nil
	}
	var users []User
	for response.Next() {
		var user User
		var publicKey string
		err = response.Scan(&user.username, &user.address, &publicKey, &user.isFriend)
		if ErrHandler(err) {
			continue
		}
		user.publicKey = decPublicKey(publicKey)
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
	privateKeyBytes := encProfileKey()
	_, err := db.Exec("INSERT INTO profiles (username, address, privateKey) VALUES ($0,$1,$2)", profile.username, profile.address, string(privateKeyBytes))
	ErrHandler(err)
}

func getProfiles() []ShowProfile {
	response, err := db.Query("SELECT id, username FROM profiles")
	if ErrHandler(err) {
		return nil
	}

	var profiles []ShowProfile
	for response.Next() {
		var profile ShowProfile
		err = response.Scan(&profile.id, &profile.username)
		if ErrHandler(err) {
			continue
		}
		profiles = append(profiles, profile)
	}
	return profiles
}

func getProfileByID(id int) (string, string, []byte) {
	response, err := db.Query("SELECT username, address, privateKey FROM profiles WHERE id=$0", id)
	if ErrHandler(err) {
		return "", "", nil
	}

	var username string
	var address string
	var privateKeyStringBytes string
	response.Next()
	err = response.Scan(&username, &address, &privateKeyStringBytes)
	if ErrHandler(err) {
		return "", "", nil
	}

	return username, address, []byte(privateKeyStringBytes)
}
