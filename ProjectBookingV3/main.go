package main

import (
	"booking/src/accountcontroller"
	"net/http"
)

func main() {

	http.HandleFunc("/account", accountcontroller.Index)
	http.HandleFunc("/account/index", accountcontroller.Index)
	http.HandleFunc("/account/login", accountcontroller.Login)
	http.HandleFunc("/account/welcome", accountcontroller.Welcome)
	http.HandleFunc("/account/logout", accountcontroller.Logout)
	http.HandleFunc("/account/book", accountcontroller.Book)
	http.HandleFunc("/account/thank", accountcontroller.Thank)

	http.ListenAndServe(":3000", nil)
}
