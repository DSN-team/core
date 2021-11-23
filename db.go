package core

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

//Private variables
var db *gorm.DB

func StartDB() {
	log.Println("Loading database...")
	db, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	ErrHandler(err)
	err := db.AutoMigrate(&User{}, &Profile{})
	ErrHandler(err)
}

func (profile *Profile) addProfile() {
	log.Println("Adding profile", profile.Username)
	db.Create(&profile)
	return
}

func getProfile(id uint, profile *Profile) {
	db.First(&profile, id)
	return
}

//TODO: generate list with username and public key
func getProfiles() (profiles []Profile) {
	result := db.Find(&profiles)
	ErrHandler(result.Error)
	return
}

func (profile *Profile) editProfile() {

}

func (profile *Profile) deleteProfile() {
	db.Delete(&profile, profile.ID)
}

func addUser(user *User) (id uint) {
	log.Println("Adding User:", user.Username)
	db.Create(&user)
	db.Last(&user)
	return user.ID
}

func (profile *Profile) getUser(id int) (user *User) {
	db.First(&user, id)
	return
}

func (profile *Profile) getUserByPublicKey(publicKey string) (user User) {
	db.Where(User{profile: profile, PublicKeyString: publicKey}).First(&user)
	return
}

func (profile *Profile) editUser(user *User) {

}

func (profile *Profile) deleteUser(user *User) {}

func (profile *Profile) addFriend(friend *Friend) (id uint) {
	db.Create(&friend)
	db.Last(&friend)
	return friend.ID
}
