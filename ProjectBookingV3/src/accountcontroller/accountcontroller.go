package accountcontroller

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"text/template"

	"github.com/gorilla/sessions"
)

var wg = sync.WaitGroup{}

var conferenceName string = "Go Conference"
var totalTickets int = 50
var totalRemainingTickets int = 50

var bookings = make([]User, 0) //struct types array

type User struct {
	userFirstName   string
	userLastName    string
	userMailAddress string
	userTickets     int
}

var store = sessions.NewCookieStore([]byte("mysession"))

func Index(response http.ResponseWriter, request *http.Request) {
	tmp, _ := template.ParseFiles("views/accountcontroller/index.html")
	tmp.Execute(response, nil)
}

func Login(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	username := request.Form.Get("username")
	password := request.Form.Get("password")

	if username == "Saiful" && password == "123" {
		session, _ := store.Get(request, "mysession")
		session.Values["username"] = username
		session.Save(request, response)
		http.Redirect(response, request, "/account/welcome", http.StatusSeeOther)
	} else {
		data := map[string]interface{}{
			"err": "Invalid",
		}
		tmp, _ := template.ParseFiles("views/accountcontroller/index.html")
		tmp.Execute(response, data)
	}
}

func Welcome(response http.ResponseWriter, request *http.Request) {
	session, _ := store.Get(request, "mysession")
	username := session.Values["username"]

	data := map[string]interface{}{
		"username":              username,
		"conferenceName":        conferenceName,
		"totalTickets":          totalTickets,
		"totalRemainingTickets": totalRemainingTickets,
	}

	tmp, _ := template.ParseFiles("views/accountcontroller/welcome.html")
	tmp.Execute(response, data)
}

func Logout(response http.ResponseWriter, request *http.Request) {
	session, _ := store.Get(request, "mysession")

	session.Options.MaxAge = -1
	session.Save(request, response)
	http.Redirect(response, request, "/account/index", http.StatusSeeOther)
}

func Book(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	firstName := request.Form.Get("firstName")
	lastName := request.Form.Get("lastName")
	email := request.Form.Get("email")
	userTicketsAsString := request.Form.Get("userTickets")
	userTickets, err := strconv.Atoi(userTicketsAsString)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(firstName, "", userTickets)

	//user validation check
	if !checkUserValidation(email) {
		session, _ := store.Get(request, "mysession")
		session.Values["firstName"] = firstName
		session.Values["lastName"] = lastName
		session.Values["email"] = email
		session.Values["userTickets"] = userTickets
		session.Values["totalRemainingTickets"] = totalRemainingTickets
		session.Save(request, response)

		bookTickets(firstName, lastName, email, userTickets)
		totalRemainingTickets = totalRemainingTickets - userTickets

		data := map[string]interface{}{
			"firstName":             firstName,
			"lastName":              lastName,
			"email":                 email,
			"userTickets":           userTickets,
			"totalRemainingTickets": totalRemainingTickets,
			"conferenceName":        conferenceName,
		}

		tmp, err := template.ParseFiles("views/accountcontroller/book.html")
		if err != nil {
			fmt.Println(err)
		}
		err = tmp.Execute(response, data)
		if err != nil {
			fmt.Println(err)
		}

		http.Redirect(response, request, "/account/thank", http.StatusSeeOther)

		wg.Add(1) //for one thread
		go sendTickets(userTickets, firstName, lastName, email)
	} else {
		data := map[string]interface{}{
			"err": "Duplicate user found",
		}
		tmp, _ := template.ParseFiles("views/accountcontroller/book.html")
		tmp.Execute(response, data)

	}

	wg.Wait()
}

func Thank(response http.ResponseWriter, request *http.Request) {
	session, _ := store.Get(request, "mysession")
	username := session.Values["username"]

	firstName, lastName, email, userTickets, totalRemainingTickets := findLastInsertedUser()
	fmt.Println("from thank you page ", firstName)

	data := map[string]interface{}{
		"firstName":             firstName,
		"lastName":              lastName,
		"email":                 email,
		"userTickets":           userTickets,
		"username":              username,
		"totalRemainingTickets": totalRemainingTickets,
	}

	tmp, _ := template.ParseFiles("views/accountcontroller/thank.html")
	tmp.Execute(response, data)
}

func bookTickets(userFirstName string, userLastName string, userMailAddress string, userTickets int) {

	var user = User{
		userFirstName:   userFirstName,
		userLastName:    userLastName,
		userMailAddress: userMailAddress,
		userTickets:     userTickets,
	}

	bookings = append(bookings, user)

	fmt.Printf("Thank you %v %v for booking %v, you will receive a confirmation mail to %v\n", userFirstName, userLastName, userTickets, userMailAddress)

	fmt.Printf("%v tickets are still remaining for %v\n", totalRemainingTickets, conferenceName)

}

// func printName() {

// 	emails := []string{}
// 	for _, booking := range bookings {

// 		emails = append(emails, booking.userMailAddress)

// 	}

// 	fmt.Printf("Total bookings are: %v\n", emails)

// }

func checkUserValidation(email string) bool {
	emails := make(map[string]int)
	for _, booking := range bookings {
		emails[booking.userMailAddress]++
	}

	for storedEmail, count := range emails {
		if count > 0 && storedEmail == email {
			fmt.Printf("Duplicate email found: %s (Count: %d)\n", email, count)
			return true
		}
	}

	return false
}

func findLastInsertedUser() (string, string, string, int, int) {
	lastIndex := len(bookings) - 1
	if lastIndex < 0 {
		return "", "", "", 0, 0
	}

	lastInsertedUser := bookings[lastIndex]
	return lastInsertedUser.userFirstName, lastInsertedUser.userLastName, lastInsertedUser.userMailAddress, lastInsertedUser.userTickets, totalRemainingTickets
}

func sendTickets(userTickets int, userFirstName string, userLastName string, userMailAddress string) {

	//time.Sleep(20 * time.Second)

	var tickets = fmt.Sprintf("%v tickets for %v %v \n to mail: %v", userTickets, userFirstName, userLastName, userMailAddress)

	fmt.Println("#################################")
	fmt.Printf("Sending tickets: \n %v\n", tickets)
	fmt.Println("#################################")

	wg.Done() // when thread is over

}
