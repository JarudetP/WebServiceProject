package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), 10)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(hash))
}
