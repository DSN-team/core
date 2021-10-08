package main

// #include <stdlib.h>
// #include <stddef.h>
// #include <stdint.h>
import "C"
import (
	"github.com/ClarkGuan/jni"
	"log"
	"reflect"
	"runtime"
	"unsafe"
)

//export Java_com_dsnteam_dsn_CoreManager_initDB
func Java_com_dsnteam_dsn_CoreManager_initDB(env uintptr, _ uintptr) {
	initDB()
}

//export Java_com_dsnteam_dsn_CoreManager_register
func Java_com_dsnteam_dsn_CoreManager_register(env uintptr, _ uintptr, usernameIn uintptr, passwordIn uintptr) (result bool) {
	username, password := string(jni.Env(env).GetStringUTF(usernameIn)), string(jni.Env(env).GetStringUTF(passwordIn))
	return register(username, password)
}

//export Java_com_dsnteam_dsn_CoreManager_login
func Java_com_dsnteam_dsn_CoreManager_login(env uintptr, _ uintptr, pos int, passwordIn uintptr) (result bool) {
	password := string(jni.Env(env).GetStringUTF(passwordIn))
	return login(password, pos)
}

//export Java_com_dsnteam_dsn_CoreManager_loadProfiles
func Java_com_dsnteam_dsn_CoreManager_loadProfiles(env uintptr, _ uintptr) int {
	return loadProfiles()
}

//export Java_com_dsnteam_dsn_CoreManager_getProfilesIds
func Java_com_dsnteam_dsn_CoreManager_getProfilesIds(env uintptr, _ uintptr) (ids uintptr) {
	ids = jni.Env(env).NewIntArray(len(profiles))
	for i := 0; i < len(profiles); i++ {
		jni.Env(env).SetIntArrayElement(ids, i, profiles[i].id)
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_getProfilesNames
func Java_com_dsnteam_dsn_CoreManager_getProfilesNames(env uintptr, _ uintptr) (usernames uintptr) {
	dataType := jni.Env(env).FindClass("Ljava/lang/String;")
	usernames = jni.Env(env).NewObjectArray(len(profiles), dataType, 0)
	for i := 0; i < len(profiles); i++ {
		jni.Env(env).SetObjectArrayElement(usernames, i, jni.Env(env).NewString(profiles[i].username))
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_getProfilePublicKey
func Java_com_dsnteam_dsn_CoreManager_getProfilePublicKey(env uintptr, _ uintptr) uintptr {
	return jni.Env(env).NewString(getProfilePublicKey())
}

//export Java_com_dsnteam_dsn_CoreManager_getProfileName
func Java_com_dsnteam_dsn_CoreManager_getProfileName(env uintptr, _ uintptr) uintptr {
	return jni.Env(env).NewString(profile.username)
}

//export Java_com_dsnteam_dsn_CoreManager_getProfileAddress
func Java_com_dsnteam_dsn_CoreManager_getProfileAddress(env uintptr, _ uintptr) uintptr {
	return jni.Env(env).NewString(profile.address)
}

//export Java_com_dsnteam_dsn_CoreManager_addFriend
func Java_com_dsnteam_dsn_CoreManager_addFriend(env uintptr, _ uintptr, addressIn uintptr, publicKeyIn uintptr) {
	address, publicKey := string(jni.Env(env).GetStringUTF(addressIn)), string(jni.Env(env).GetStringUTF(publicKeyIn))
	addFriend(address, publicKey)
}

//export Java_com_dsnteam_dsn_CoreManager_loadFriends
func Java_com_dsnteam_dsn_CoreManager_loadFriends(env uintptr, _ uintptr) int {
	return loadFriends()
}

//export Java_com_dsnteam_dsn_CoreManager_getFriendsIds
func Java_com_dsnteam_dsn_CoreManager_getFriendsIds(env uintptr, _ uintptr) (ids uintptr) {
	ids = jni.Env(env).NewIntArray(len(friends))
	for i := 0; i < len(friends); i++ {
		jni.Env(env).SetIntArrayElement(ids, i, friends[i].id)
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_getFriendsNames
func Java_com_dsnteam_dsn_CoreManager_getFriendsNames(env uintptr, _ uintptr) (usernames uintptr) {
	//friends = getFriends()
	dataType := jni.Env(env).FindClass("Ljava/lang/String;")
	usernames = jni.Env(env).NewObjectArray(len(friends), dataType, 0)
	for i := 0; i < len(friends); i++ {
		jni.Env(env).SetObjectArrayElement(usernames, i, jni.Env(env).NewString(friends[i].username))
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_getFriendsAddresses
func Java_com_dsnteam_dsn_CoreManager_getFriendsAddresses(env uintptr, _ uintptr) (address uintptr) {
	//friends = getFriends()
	dataType := jni.Env(env).FindClass("Ljava/lang/String;")
	address = jni.Env(env).NewObjectArray(len(friends), dataType, 0)
	for i := 0; i < len(friends); i++ {
		jni.Env(env).SetObjectArrayElement(address, i, jni.Env(env).NewString(friends[i].address))
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_getFriendsPublicKeys
func Java_com_dsnteam_dsn_CoreManager_getFriendsPublicKeys(env uintptr, _ uintptr) (publicKey uintptr) {
	//friends = getFriends()
	dataType := jni.Env(env).FindClass("Ljava/lang/String;")
	publicKey = jni.Env(env).NewObjectArray(len(friends), dataType, 0)
	for i := 0; i < len(friends); i++ {
		jni.Env(env).SetObjectArrayElement(publicKey, i, jni.Env(env).NewString(encPublicKey(marshalPublicKey(friends[i].publicKey))))
	}
	return
}

//export Java_com_dsnteam_dsn_CoreManager_connectToFriends
func Java_com_dsnteam_dsn_CoreManager_connectToFriends(env uintptr, _ uintptr) {
	connectToFriends()
}

//export Java_com_dsnteam_dsn_CoreManager_connectToFriend
func Java_com_dsnteam_dsn_CoreManager_connectToFriend(env uintptr, _ uintptr, userId int) {
	connectToFriend(userId)
}

//Инициализировать структуры и подключение

//export Java_com_dsnteam_dsn_CoreManager_runServer
func Java_com_dsnteam_dsn_CoreManager_runServer(env uintptr, _ uintptr, addressIn uintptr) {
	address := string(jni.Env(env).GetStringUTF(addressIn))
	println("env run:", env)
	if env != 0 {
		workingVM, _ = jni.Env(env).GetJavaVM()
	}
	runServer(address)
}

//export Java_com_dsnteam_dsn_CoreManager_setCallBackBuffer
func Java_com_dsnteam_dsn_CoreManager_setCallBackBuffer(env uintptr, _ uintptr, jniBuffer uintptr) {
	callBackBufferPtr = jni.Env(env).GetDirectBufferAddress(jniBuffer)
	callBackBufferCap = jni.Env(env).GetDirectBufferCapacity(jniBuffer)
}

//export Java_com_dsnteam_dsn_CoreManager_writeBytes
func Java_com_dsnteam_dsn_CoreManager_writeBytes(env uintptr, _ uintptr, inBuffer uintptr, lenIn int, userId int) {
	log.Println("env write:", env)
	defer runtime.KeepAlive(dataStrInput.io)
	point, size := jni.Env(env).GetDirectBufferAddress(inBuffer), jni.Env(env).GetDirectBufferCapacity(inBuffer)

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&dataStrInput.io))
	sh.Data = uintptr(point)
	sh.Len = lenIn
	sh.Cap = size

	writeBytes(userId, lenIn)
}
