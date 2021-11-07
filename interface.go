package core

import (
	"fmt"
	"log"
)

var UpdateUI = func(int, int) {}

func (cur *Profile) Register(username, password, address string) bool {
	key := genProfileKey()
	if key == nil {
		return false
	}
	cur.Username, cur.Password, cur.Address, cur.PrivateKey = username, password, address, key
	log.Println(cur)
	cur.Id = addProfile(cur)
	return true
}

func (cur *Profile) Login(password string, pos int) (result bool) {
	var privateKeyEncBytes []byte
	cur.Id = Profiles[pos].Id
	cur.Username, cur.Address, privateKeyEncBytes = getProfileByID(Profiles[pos].Id)
	fmt.Println("privateKeyEncBytes: ", privateKeyEncBytes)
	if privateKeyEncBytes == nil {
		return false
	}
	result = cur.decProfileKey(privateKeyEncBytes, password)
	fmt.Println("Login status:", result)
	return
}

func UsernamePos(username string) int {
	profiles := getProfiles()
	pos := -1
	for i := 0; i < len(profiles); i++ {
		if profiles[i].Username == username {
			pos = i
			break
		}
	}
	return pos
}

//TODO Deprecated
//func (cur *Profile) WriteDataRequest(userId, lenIn int) {
//	var con net.Conn
//	if _, ok := cur.Connections.Load(userId); !ok {
//		log.Println("Not connected to:", userId)
//		return
//	}
//	value, _ := cur.Connections.Load(userId)
//	con = value.(net.Conn)
//	runtime.KeepAlive(cur.DataStrInput.Io)
//	log.Println("writing to:", con.RemoteAddr())
//
//	log.Println("input:", cur.DataStrInput.Io)
//	println("input str:", string(cur.DataStrInput.Io))
//
//	switch err {
//	case nil:
//		bytes := BuildRequest(RequestData, uint64(lenIn), cur.DataStrInput.Io[0:lenIn])
//		println("ClientSend:", bytes, " count:", lenIn)
//		if _, err = con.Write(bytes); err != nil {
//			log.Printf("failed to send the client request: %v\n", err)
//		}
//
//	case io.EOF:
//		log.Println("client closed the connection")
//		return
//	default:
//		log.Printf("client error: %v\n", err)
//		return
//	}
//}
