package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/gustavNdamukong/hotel-bookings/internal/models"
	mail "github.com/xhit/go-simple-mail"
)

func listenForMail() {
	//code in a channel needs to run in the background (asynchronously) for it to
	// fulfil the purpose of a channel. It should never stop the app from running
	// in go you do that by prefixing the code execution or call to any func with
	// the 'go' keyword eg we can also do that to execute an anonymous func like so:

	// NOTES: Here's how to create an anonymous func in go
	go func() {
		// This for loop means that we will be listening to this channel indefinitely
		for {
			msg := <-app.MailChan
			sendMsg(msg)
		}
	}()

}

// sendMsg is a custom func. It doesn't return anything. It just sends email
func sendMsg(m models.MailData) {
	server := mail.NewSMTPClient()

	//tell it where the mail server is. We will install a dummy mail server on our machine shortly
	server.Host = "localhost"

	//tell it which port the mail server will be exposing
	// our dummy test server which we will install will use port 1025. Real mail servers use ports
	// like port 25, 587, or 465
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second //(10 secs)
	server.SendTimeout = 10 * time.Second    //(10 secs)

	//In production you will likely need more settings like:
	/*server.Username ...
	server.Password ...
	server.Encryption ...
	etc */

	client, err := server.Connect()
	if err != nil {
		errorLog.Println(err)
	}

	//Construct your email message in a format that our client understands. We do this using m
	// mail has a method NewMSG() below which returns a struct which is a pointer to email & a new email msg
	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)

	//check if a template has ben specified
	if m.Template == "" {
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		// grab the template (from /email-templates)
		// NOTES: How to read from files from disk. It uses the input/output utility (ioutil)
		data, err := ioutil.ReadFile(fmt.Sprintf("./email-templates/%s", m.Template))
		if err != nil {
			app.ErrorLog.Println(err)
		}

		//cast that template data to a string from the array of bytes returned by the ioutil's ReadFile() func
		// NOTES: Here is how to quickly cast data into a string-using the string() helper func
		mailTemplate := string(data)

		//now we need to substitute the data placeholders in the template files. fmt.Sprintf()
		//	will not cut is this time
		// we will need a proper (dedicated) string function strings.Replace(). The last argument
		//	1 tells Go to do the replacement just once.
		msgToSend := strings.Replace(mailTemplate, "[%body%]", m.Content, 1)
		email.SetBody(mail.TextHTML, msgToSend)
	}

	err = email.Send(client)
	if err != nil {
		errorLog.Println(err)
	} else {
		log.Println("Email sent!")
	}
}
