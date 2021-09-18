package main

import "log"

var err error

func ErrHandler(err error) bool {
	if err != nil {
		log.Fatalln(err)
		return true
	}
	return false
}
