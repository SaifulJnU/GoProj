package accountcontroller

import (
	"fmt"
	"net/http"
	"net/smtp"
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
	///password := request.Form.Get("password")

	if true {
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

		totalRemainingTickets = totalRemainingTickets - userTickets
		bookTickets(firstName, lastName, email, userTickets)

		session.Save(request, response)

		// bookTickets(firstName, lastName, email, userTickets)
		// totalRemainingTickets = totalRemainingTickets - userTickets

		// data := map[string]interface{}{
		// 	"firstName":             firstName,
		// 	"lastName":              lastName,
		// 	"email":                 email,
		// 	"userTickets":           userTickets,
		// 	"totalRemainingTickets": totalRemainingTickets,
		// 	"conferenceName":        conferenceName,
		// }

		// tmp, err := template.ParseFiles("views/accountcontroller/booking.html")
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// err = tmp.Execute(response, data)
		// if err != nil {
		// 	fmt.Println(err)
		// }

		http.Redirect(response, request, "/account/thank", http.StatusFound)

		wg.Add(1) //for one thread
		go sendTickets(userTickets, firstName, lastName, email)

	} else {
		data := map[string]interface{}{
			"err": "Duplicate user found",
		}
		tmp, _ := template.ParseFiles("views/accountcontroller/booking.html")
		tmp.Execute(response, data)

	}

	wg.Wait()
}

func Booking(response http.ResponseWriter, request *http.Request) {
	tmp, _ := template.ParseFiles("views/accountcontroller/booking.html")
	tmp.Execute(response, nil)
}

func Thank(response http.ResponseWriter, request *http.Request) {
	session, _ := store.Get(request, "mysession")
	username := session.Values["username"]

	// data := map[string]interface{}{
	// 	"firstName":             firstName,
	// 	"lastName":              lastName,
	// 	"email":                 email,
	// 	"userTickets":           userTickets,
	// 	"totalRemainingTickets": totalRemainingTickets,
	// 	"conferenceName":        conferenceName,
	// }

	// tmp, err := template.ParseFiles("views/accountcontroller/booking.html")

	firstName, lastName, email, userTickets, totalRemainingTickets := findLastInsertedUser()
	//fmt.Println("from thank you page ", firstName)

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

	sendRealmail(firstName, lastName, email, userTickets, totalRemainingTickets)

}

func bookTickets(userFirstName string, userLastName string, userMailAddress string, userTickets int) {

	var user = User{
		userFirstName:   userFirstName,
		userLastName:    userLastName,
		userMailAddress: userMailAddress,
		userTickets:     userTickets,
	}

	bookings = append(bookings, user)

	// fmt.Printf("Thank you %v %v for booking %v, you will receive a confirmation mail to %v\n", userFirstName, userLastName, userTickets, userMailAddress)

	// fmt.Printf("%v tickets are still remaining for %v\n", totalRemainingTickets, conferenceName)

}

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

	//time.Sleep(2 * time.Second)

	var tickets = fmt.Sprintf("%v tickets for %v %v \n to mail: %v", userTickets, userFirstName, userLastName, userMailAddress)

	fmt.Println("#################################")
	fmt.Printf("Sending tickets: \n %v\n", tickets)
	fmt.Println("#################################")

	wg.Done() // when thread is over

}

func sendRealmail(firstName string, lastName string, email string, userTickets int, totalReamainingTickets int) {

	// Sender data
	from := "developer.saiful98@gmail.com"
	password := "xgcfbligriyajshc"

	// Receiver data
	toEmail := "saiful.cse98@gmail.com"
	to := []string{toEmail}

	host := "smtp.gmail.com"
	port := "587"
	address := host + ":" + port

	// Message
	subject := "Subject: Confirmation mail for Go conference ticket booking\r\n"

	body := fmt.Sprintf("You have successfully booked %v tickets for %v %v \n to mail: %v \n We still have %v tickets left for the conference", userTickets, firstName, lastName, email, totalReamainingTickets)

	message := []byte(subject + "\r\n" + body)

	auth := smtp.PlainAuth("", from, password, host)
	err := smtp.SendMail(address, auth, from, to, message)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Please check your mail for the confirmation.")
}
