package core

import (
	"fmt"
	//"database/sql"
	//_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

var db *gorm.DB

func StartDB() {
	fmt.Println("Loading database...")
	db, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	ErrHandler(err)
	err := db.AutoMigrate(&User{}, &Profile{}, &UserRequest{})
	ErrHandler(err)
}

func addProfile(profile *Profile) {
	fmt.Println("Adding profile", profile.Username)
	result := db.Create(&profile)
	log.Println(result.RowsAffected)
	return
}

func getProfiles() []ShowProfile {
	fmt.Println("Getting profiles...")
	var outputProfiles []ShowProfile
	db.Model(&Profile{}).Find(&outputProfiles)
	return outputProfiles
}

func getProfileByID(id uint) (profile *Profile) {
	fmt.Println("Getting profile by ID", id)
	db.First(&profile, id)
	return
}

func (cur *Profile) addUser(user *User) {
	fmt.Println("Adding User:", user.Username)
	var search User
	db.Where(user).First(&search)
	if search.ID == 0 {
		db.Create(&user)
		db.Last(&user)
	} else {
		log.Println("User already exists", user.ID)
		user.ID = search.ID
	}
	return
}

func (cur *Profile) editUser(id int, user *User) {
	fmt.Println("Editing user", id)
	db.First(&user, id)
}

func (cur *Profile) GetUser(id uint) (user User) {
	fmt.Println("Getting user", id)
	db.Where(User{ProfileID: cur.ID}).First(&user, id)
	return
}

func (cur *Profile) getUserByUsername(username string) (user User) {
	fmt.Println("Getting user by username", username)
	db.Where(&User{ProfileID: cur.ID, Username: username}).First(&user)
	publicKey := UnmarshalPublicKey(DecodeKey(user.PublicKeyString))
	user.PublicKey = &publicKey
	return
}

func (cur *Profile) getUserByPublicKey(publicKeyString string) (user User) {
	fmt.Println("Getting user by public key")
	db.Where(User{PublicKeyString: publicKeyString}).First(&user)
	publicKey := UnmarshalPublicKey(DecodeKey(user.PublicKeyString))
	user.PublicKey = &publicKey
	return
}

func (cur *Profile) addFriendRequest(userID uint, direction int) {
	fmt.Println("Adding friend request: friend userID =", userID)
	var request UserRequest
	db.Where(&UserRequest{ProfileID: cur.ID, UserID: userID, Direction: direction}).Find(&request)
	if request.ID == 0 {
		db.Create(&UserRequest{ProfileID: cur.ID, UserID: userID, Direction: direction, Status: 1})
	} else {
		log.Println("Request already exists")
	}
}

func (cur *Profile) AcceptFriendRequest(request *UserRequest) {
	fmt.Println("Accept friend request:", request.UserID)
	user := cur.GetUser(request.UserID)
	user.IsFriend = true
	db.Save(&user)
	request.Status = 2
	db.Save(&request)
}

func (cur *Profile) RejectFriendRequest(request *UserRequest) {
	fmt.Println("Rejected friend request:", request.UserID)
	user := cur.GetUser(request.UserID)
	user.IsFriend = false
	db.Save(&user)
	request.Status = 2
	db.Save(&request)
}

func (cur *Profile) DeleteFriendRequest(request *UserRequest) {
	fmt.Println("Deleted friend request:", request.UserID)
	db.Delete(&request)
}

func (cur *Profile) getFriendRequestsIn() (requests []UserRequest) {
	fmt.Println("Loading friend incoming requests")
	db.Where(&UserRequest{ProfileID: cur.ID, Direction: 0, Status: 1}).Find(&requests)
	return requests
}

func (cur *Profile) getFriendRequestsOut() (requests []UserRequest) {
	fmt.Println("Loading friend outgoing requests")
	db.Where(&UserRequest{ProfileID: cur.ID, Direction: 1, Status: 1}).Find(&requests)
	return requests
}

func (cur *Profile) searchFriendRequest(id uint) bool {
	fmt.Println("Searching friend request:", id)
	var request UserRequest
	db.Where(UserRequest{ProfileID: cur.ID, UserID: id}, "friendRequests").Find(&request)
	if request.ID == 0 {
		return false
	} else {
		return true
	}
}

func (cur *Profile) getFriends() []User {
	fmt.Println("Getting friends")
	var users []User
	db.Where(&User{ProfileID: cur.ID, IsFriend: true}).Find(&users)
	return users
}
