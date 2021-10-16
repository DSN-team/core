package core

import "log"

var err error

func ErrHandler(err error) bool {
	if err != nil {
		log.Println(err)
		return true
	}
	return false
}
