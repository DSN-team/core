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
	err := db.AutoMigrate(&User{}, &Profile{}, &UserRequest{})
	ErrHandler(err)
}

func (cur *Profile) addFriendRequest(userID uint, direction int) {
	log.Println("Adding friend request: friend userID =", userID)
	db.Create(&UserRequest{ProfileID: cur.ID, UserID: userID, Direction: direction})
}

func (cur *Profile) GetFriendRequests() (requests []UserRequest) {
	db.Where("profile_id", cur.ID).Find(&requests)
	return requests
}

func (cur *Profile) searchFriendRequest(id uint) bool {
	var request UserRequest
	db.Where(UserRequest{ProfileID: cur.ID, UserID: id}, "friendRequests").Find(&request)
	if request.ID == 0 {
		return false
	} else {
		return true
	}
}

func (cur *Profile) searchUser(username string) (user User) {
	db.Where(&User{ProfileID: cur.ID, Username: username}).First(&user)
	publicKey := UnmarshalPublicKey(DecodeKey(user.PublicKeyString))
	user.PublicKey = &publicKey
	return
}

func (cur *Profile) addUser(user *User) {
	log.Println("Adding User:", user.Username)
	db.Create(&user)
	db.Last(&user)
	return
}

func (cur *Profile) editUser(id int, user *User) {
	log.Println("Editing user", id)
	//db.Exec("UPDATE users SET address = $1, public_key = $2 WHERE id = $0 and profile_id = $3", id, user.Address, EncodeKey(MarshalPublicKey(user.PublicKey)), cur.User.Id)
	db.First(&user, id)
}

func (cur *Profile) getFriends() []User {
	var users []User
	db.Where(&User{ProfileID: cur.ID, IsFriend: true}).Find(&users)
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

func (cur *Profile) getUserByPublicKey(publicKeyString string) (user User) {
	db.Where(User{PublicKeyString: publicKeyString}).First(&user)
	publicKey := UnmarshalPublicKey(DecodeKey(user.PublicKeyString))
	user.PublicKey = &publicKey
	return
}
