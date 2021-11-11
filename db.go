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
	_, err := db.Exec("create table if not exists users (id integer not null constraint users_pk primary key autoincrement, profile_id integer not null, username text, address text not null, public_key text not null, is_friend integer default 0)")
	ErrHandler(err)
	_, err = db.Exec("create unique index if not exists users_id_uindex on users (id)")
	ErrHandler(err)
}

func createProfilesTable() {
	log.Println("creating Profiles table")
	_, err := db.Exec("create table if not exists profiles (id integer not null constraint profiles_pk primary key autoincrement, username text, address text, private_key text)")
	ErrHandler(err)
	_, err = db.Exec("create unique index if not exists profiles_id_uindex on profiles (id)")
	ErrHandler(err)
}

func createFriendRequestsTable() {
	log.Println("creating friends requests table")
	_, err := db.Exec("create table if not exists friends_requests (id integer not null constraint friends_requests_pk primary key autoincrement, profile_id integer not null, friend_id integer not null)")
	ErrHandler(err)
	_, err = db.Exec("create unique index if not exists friends_requests_id_uindex on friends_requests (id)")
	ErrHandler(err)
}

func StartDB() {
	log.Println("initing db")
	db, err = sql.Open("sqlite3", "data.db")
	ErrHandler(err)
	_, err := db.Begin()
	ErrHandler(err)
	createProfilesTable()
	createUsersTable()
	createFriendRequestsTable()
}

func (cur *Profile) addFriendRequest(id int) {
	log.Println("Adding friend request: friend id =", id)
	_, err := db.Exec("INSERT INTO (profile_id, friend_id) VALUES ($0,$1)", cur.ThisUser.Id, id)
	ErrHandler(err)
}

//todo array of ids
func (cur *Profile) getFriendRequests() {}

func (cur *Profile) searchUser(username string) (id int) {
	log.Println("searching user, ", username)
	rows, err := db.Query("SELECT id FROM users WHERE username = $0 limit 1", username)
	id = -1
	if ErrHandler(err) {
		return
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		ErrHandler(err)
	}(rows)
	for rows.Next() {
		err = rows.Scan(&id)
		if ErrHandler(err) {
			return
		}
	}
	return
}

func (cur *Profile) addUser(user User) (id int) {
	log.Println("Adding user", user.Username)
	_, err := db.Exec("INSERT INTO users (profile_id,username,address,public_key,is_friend) VALUES ($0,$1,$2,$3,$5)", cur.ThisUser.Id, user.Username, user.Address, EncPublicKey(MarshalPublicKey(user.PublicKey)), user.IsFriend)
	ErrHandler(err)
	rows, err := db.Query("SELECT id FROM users ORDER BY column DESC LIMIT 1")

	defer func(rows *sql.Rows) {
		err := rows.Close()
		ErrHandler(err)
	}(rows)
	id = -1
	rows.Next()
	err = rows.Scan(&id)
	return
}

func (cur *Profile) editUser(id int, user User) {
	log.Println("Editing user", id)
	_, err := db.Exec("UPDATE users SET address = $1, public_key = $2 WHERE id = $0", id, user.Address, EncPublicKey(MarshalPublicKey(user.PublicKey)))
	if ErrHandler(err) {
		return
	}
}

func (cur *Profile) getFriends() []User {
	log.Println("Getting Friends")
	rows, err := db.Query("SELECT id, username, address, public_key, is_friend FROM users WHERE is_friend = 1 and profile_id = $0", cur.ThisUser.Id)
	if ErrHandler(err) {
		return nil
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		ErrHandler(err)
	}(rows)
	var users []User
	number := 0
	for rows.Next() {
		var user User
		var publicKey string
		err = rows.Scan(&user.Id, &user.Username, &user.Address, &publicKey, &user.IsFriend)
		if ErrHandler(err) {
			continue
		}
		decryptedPublicKey := UnmarshalPublicKey(DecPublicKey(publicKey))
		user.PublicKey = &decryptedPublicKey
		users = append(users, user)
		cur.FriendsIDXs.Store(number, user.Id)
		number++
	}
	return users
}

func addProfile(profile *Profile) (id int) {
	log.Println("Adding Profile", profile.ThisUser.Username)
	privateKeyBytes := profile.encProfileKey()
	_, err := db.Exec("INSERT INTO profiles (username, address, private_key) VALUES ($0,$1,$2)", profile.ThisUser.Username, profile.ThisUser.Address, string(privateKeyBytes))
	ErrHandler(err)
	rows, err := db.Query("SELECT id FROM profiles ORDER BY id DESC LIMIT 1")

	defer func(rows *sql.Rows) {
		err := rows.Close()
		ErrHandler(err)
	}(rows)
	id = -1
	rows.Next()
	err = rows.Scan(&id)
	return
}

func getProfiles() []ShowProfile {
	log.Println("Getting Profiles")

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
		err = rows.Scan(&profile.Id, &profile.Username)
		if ErrHandler(err) {
			continue
		}
		profiles = append(profiles, profile)
	}
	return profiles
}

func getProfileByID(id int) (string, string, []byte) {
	log.Println("Getting Profile by ID", id)
	rows, err := db.Query("SELECT username, address, private_key FROM profiles WHERE id=$0", id)
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
	log.Println("Getting Profile by key")
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
