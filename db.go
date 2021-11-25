package core

import (
	//"database/sql"
	//_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

var db *gorm.DB

func StartDB() {
	log.Println("Loading database...")
	db, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	ErrHandler(err)
	err := db.AutoMigrate(&User{}, &Profile{})
	ErrHandler(err)
}

func (cur *Profile) addFriendRequest(userID uint, direction int) {
	log.Println("Adding friend request: friend userID =", userID)
	db.Create(&UserRequest{ProfileID: cur.ID, UserID: userID, Direction: direction})
}

func (cur *Profile) GetFriendRequests() []UserRequest {
	var requests []UserRequest
	res := db.Find(&requests)
	ErrHandler(res.Error)
	return requests
}

func (cur *Profile) searchFriendRequest(id uint) bool {
	var requests []UserRequest
	db.Where(UserRequest{ProfileID: cur.ID, UserID: id}, "friendRequests").Find(&requests)
	if requests == nil {
		return false
	} else {
		return true
	}
}

func (cur *Profile) searchUser(username string) (user User) {
	db.Where(&User{Username: username}, "username").First(&user)
	return
}

func (cur *Profile) addUser(user *User) {
	log.Println("Adding User:", user.Username)
	db.Create(&user)
	db.Last(&user)
	return
}

func (cur *Profile) editUser(id int, user User) {
	log.Println("Editing user", id)
	//db.Exec("UPDATE users SET address = $1, public_key = $2 WHERE id = $0 and profile_id = $3", id, user.Address, EncPublicKey(MarshalPublicKey(user.PublicKey)), cur.User.Id)
	db.First(&user, id)
}

func (cur *Profile) getFriends() []User {
	var users []User
	db.Where("is_friend = true").Find(&users)
	return users
}

func addProfile(profile *Profile) {
	log.Println("Adding profile", profile.Username)
	result := db.Create(&profile)
	log.Println(result.RowsAffected)
	return
}

func getProfiles() []ShowProfile {
	var outputProfiles []ShowProfile
	db.Model(&Profile{}).Find(&outputProfiles)
	return outputProfiles
}

func getProfileByID(id uint) (profile *Profile) {
	db.First(&profile, id)
	return
}

func (cur *Profile) getUserByPublicKey(publicKey string) (user User) {
	db.Where(User{PublicKeyString: publicKey}).First(&user)
	return
}
