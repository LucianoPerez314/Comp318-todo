package main

import (
	"fmt"
	"net/http"

	"github.com/LucianoPerez314/Comp318-todo/todo"
)

func main() {
	//Create todo Multiplexor
	todos := todo.New()
	//Initialize and start server
	var srv http.Server
	srv.Addr = "localhost:5318"
	srv.Handler = todos
	err := srv.ListenAndServe()
	fmt.Println(err)
}
