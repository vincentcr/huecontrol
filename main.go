package main

import (
	"fmt"

	"github.com/vincentcr/huecontrol/services"
)

func main() {
	svc, err := services.New("dev")
	if err != nil {
		panic(err.Error())
	}

	u0, err := svc.Users.Create("foo@bar.com", "123333")
	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println(u0)
	}

	u, err := svc.Users.AuthenticateWithPassword("vincentcr@gmail.com", "abcdefg")
	if err != nil {
		fmt.Println("ERR!!!!", err.Error())
	} else {
		fmt.Println(u)
	}
}
