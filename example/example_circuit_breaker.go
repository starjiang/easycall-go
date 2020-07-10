package main

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/starjiang/easycall-go"
)

var count = 0

func getUserName() error {
	if count > 15 {
		fmt.Println("NO ERROR")
		return nil
	}
	count++
	fmt.Println("ERROR")
	return errors.New("invalid params")
}

func fail() error {
	fmt.Println("FAIL")
	return nil
}

func main() {
	flag.Parse()
	for i := 0; i < 200; i++ {
		err := easycall.CbCall("hello", getUserName, fail)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Second)
	}
}
