package core

import (
	//"database/sql"
	//_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

//Private variables
//var db *sql.DB
var db *gorm.DB

//var dbWG *sync.WaitGroup

func createUsersTable() {
	/*
		log.Println("creating users table")
		_, err := db.Exec("create table if not exists users (id integer not null constraint users_pk primary key autoincrement, profile_id integer not null, username text, address text not null, public_key text not null, is_friend integer default 0)")
		ErrHandler(err)
		_, err = db.Exec("create unique index if not exists users_id_uindex on users (id)")
		ErrHandler(err)

	*/
}

func createProfilesTable() {
	/*
		log.Println("creating Profiles table")
		_, err := db.Exec("create table if not exists profiles (id integer not null constraint profiles_pk primary key autoincrement, username text, address text, private_key text)")
		ErrHandler(err)
		_, err = db.Exec("create unique index if not exists profiles_id_uindex on profiles (id)")
		ErrHandler(err)
	*/

}

func createFriendRequestsTable() {
	/*
		log.Println("creating friends requests table")
		_, err := db.Exec("create table if not exists friends_requests (id integer not null constraint friends_requests_pk primary key autoincrement, profile_id integer not null, friend_id integer not null, direction integer not null)")
		ErrHandler(err)
		_, err = db.Exec("create unique index if not exists friends_requests_id_uindex on friends_requests (id)")
		ErrHandler(err)

	*/
}

func StartDB() {
	/*log.Println("initing db")
	db, err = sql.Open("sqlite3", "data.db")
	ErrHandler(err)
	_, err := db.Begin()
	ErrHandler(err)
	createProfilesTable()
	createUsersTable()
	createFriendRequestsTable()
	*/

	log.Println("Loading database...")
	db, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	ErrHandler(err)
	err := db.AutoMigrate(&User{}, &Profile{})
	ErrHandler(err)
}

func (cur *Profile) addFriendRequest(id int, direction int) {
	log.Println("Adding friend request: friend id =", id)
	//_, err := db.Exec("INSERT INTO friends_requests (profile_id, friend_id, direction) VALUES ($0,$1,$2)", cur.User.Id, id, direction)
	//ErrHandler(err)
	db.Create(cur.FriendRequests[len(cur.FriendRequests)-1])
}

func (cur *Profile) GetFriendRequests() []FriendRequest {
	/*
		var friendRequests []FriendRequest
		rows, err := db.Query("SELECT users.public_key, users.username FROM friends_requests JOIN users ON friends_requests.friend_id = users.id WHERE profile_id = $0 DESC", cur.User.Id)
		if ErrHandler(err) {
			return nil
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			ErrHandler(err)
		}(rows)
		for rows.Next() {
			friendRequest := FriendRequest{}
			err = rows.Scan(&friendRequest.PublicKey, &friendRequest.Username)
			if ErrHandler(err) {
				return nil
			}
			friendRequests = append(friendRequests, friendRequest)
		}
		return friendRequests
	*/
	var requests []FriendRequest
	res := db.Find(&requests)
	ErrHandler(res.Error)
	return requests
}

func (cur *Profile) searchFriendRequest(id int) bool {
	/*
		log.Println("searching friend request: friend id =", id)
		rows, err := db.Query("SELECT id FROM friends_requests WHERE friend_id = $0 and profile_id = $1 limit 1", id, cur.User.Id)
		id = -1
		if ErrHandler(err) {
			return false
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			ErrHandler(err)
		}(rows)
		rows.Next()
		err = rows.Scan(&id)
		if ErrHandler(err) {
			return false
		} else {
			if id == -1 {
				return false
			} else {
				return true
			}
		}
	*/
	var requests []FriendRequest
	db.Where(FriendRequest{ID: uint(id)}, "friendRequests").Find(&requests)
	if requests == nil {
		return false
	} else {
		return true
	}
}

func (cur *Profile) searchUser(username string) (id int) {
	/*
		log.Println("searching user,", username)
		rows, err := db.Query("SELECT id FROM users WHERE username = $0 and profile_id = $1 limit 1", username, cur.User.Id)
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
	*/
	var users []User
	db.Where(&User{Username: username}, "username").Find(&users)
	if users != nil {
		users[0].Id = int(users[0].ID)
		return users[0].Id
	} else {
		return -1
	}
}

func (cur *Profile) addUser(user User) (id int) {
	/*log.Println("Adding user, username:", user.Username, "address:", user.Address, "public key:", user.PublicKey)
	_, err = db.Exec("INSERT INTO users (profile_id,username,address,public_key,is_friend) VALUES ($0,$1,$2,$3,$5)", cur.User.Id, user.Username, user.Address, EncPublicKey(MarshalPublicKey(user.PublicKey)), user.IsFriend)
	ErrHandler(err)
	rows, err := db.Query("SELECT id FROM users ORDER BY id DESC LIMIT 1")
	ErrHandler(err)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		ErrHandler(err)
	}(rows)
	id = -1
	rows.Next()
	err = rows.Scan(&id)
	return */
	log.Println("Adding User:", user.Username)
	db.Create(&user)
	db.Last(&user)
	user.Id = int(user.ID)
	return user.Id
}

func (cur *Profile) editUser(id int, user User) {
	log.Println("Editing user", id)
	//db.Exec("UPDATE users SET address = $1, public_key = $2 WHERE id = $0 and profile_id = $3", id, user.Address, EncPublicKey(MarshalPublicKey(user.PublicKey)), cur.User.Id)
	db.First(&user, id)
}

func (cur *Profile) getFriends() []User {
	/*
		log.Println("Getting Friends")
		rows, err := db.Query("SELECT id, username, address, public_key, is_friend FROM users WHERE is_friend = 1 and profile_id = $0", cur.User.Id)
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
	*/
	var users []User
	result := db.Find(&users)
	ErrHandler(result.Error)
	return users
}

func addProfile(profile *Profile) (id int) {
	/*log.Println("Adding Profile", profile.User.Username)
	privateKeyBytes := profile.encProfileKey()
	_, err = db.Exec("INSERT INTO profiles (username, address, private_key) VALUES ($0,$1,$2)", profile.User.Username, profile.User.Address, string(privateKeyBytes))
	ErrHandler(err)
	rows, err := db.Query("SELECT id FROM profiles ORDER BY id DESC LIMIT 1")
	ErrHandler(err)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		ErrHandler(err)
	}(rows)
	id = -1
	rows.Next()
	err = rows.Scan(&id)
	return*/
	log.Println("Adding profile", profile.User.Username)
	db.Create(&profile)
	return
}

func getProfiles() []ShowProfile {
	/*
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
	*/
	var output_profiles []ShowProfile
	db.Model(&Profile{}).Find(output_profiles)
	return output_profiles
}

func getProfileByID(id int) (string, string, []byte) {
	/*
		log.Println("Getting Profile by Id", id)
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
	*/
	var profile Profile
	db.First(&profile, id)
	return profile.User.Username, profile.User.Address, []byte(profile.User.PublicKeyStr)
}

func (cur *Profile) getUserByPublicKey(publicKey string) int {
	/*log.Println("Getting Profile by key")
	rows, err := db.Query("SELECT id FROM users WHERE public_key=$0 and profile_id=$1", publicKey, cur.User.Id)
	if ErrHandler(err) {
		return 0
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		ErrHandler(err)
	}(rows)

	id := -1
	rows.Next()
	err = rows.Scan(&id)
	if ErrHandler(err) {
		return 0
	}
	return id*/
	var user User
	db.Where(User{PublicKeyStr: publicKey}).First(&user)
	return user.Id
}
