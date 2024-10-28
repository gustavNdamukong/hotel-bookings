package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gustavNdamukong/hotel-bookings/internal/config"
	"github.com/gustavNdamukong/hotel-bookings/internal/driver"
	"github.com/gustavNdamukong/hotel-bookings/internal/handlers"
	"github.com/gustavNdamukong/hotel-bookings/internal/helpers"
	"github.com/gustavNdamukong/hotel-bookings/internal/models"
	"github.com/gustavNdamukong/hotel-bookings/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// main is the main function
func main() {
	// NOTES: add to debug notes that the equivalent of dump & die in go is
	// log.Fatal(err) coz it will abort the app execution & log the error. Remember to import log above though
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()
	defer close(app.MailChan)

	fmt.Println("Starting mail listener")

	listenForMail()
	/* We dont wanna be sending an email every time we start our server, just yet
	msg := models.MailData{
		To:      "john@do.ca",
		From:    "me@here.ca",
		Subject: "Some subject",
		Content: "",
	}

	app.MailChan <- msg
	*/

	//----------------------Sending email with the standard library------------------------
	/*
		// NOTES: Sending emails in Go. Its poss with standard lib, but not ideal. Here's how
		// write code so when the program starts it sends an email
		from := "me@here.com"
		// mail server dredencials (mandatory)
		// smtp.PlainAuth() takes your mail server ID, a username, your smailserver password if any, and hostname
		auth := smtp.PlainAuth("", from, "", "localhost")
		// create and send msg in one step
		//smtp.SendMAil() requires an address-in this case for ur localhost server & port num that will be
		//	applicable one u've installed the nec package, the auth u defined above, the from email address,
		//	the recipient's address (here we just use a slice of strings []string{"you@there.com"})
		//	finally, it accepts the email msg content, which we pass here as a slice of bytes []byte("Hello world")
		err = smtp.SendMAil("localhost:1025", auth, from, []string{"you@there.com"}, []byte("Hello world"))
		if err != nil {
			log.Println(err)
		}
	*/

	/*
		 -You will get an error: 'package smtp is not in std (/usr/local/go/src/smtp)'. That's coz
		  as of Go 1.17, the net/smtp package is deprecated and has been removed from the standard library.
		  To fix this issue, you'll need to use an external package to send emails. One popular package
		  is gomail (github.com/go-gomail/gomail).

		-You used to be able top visit your test mail server which is on: 'http://localhost:8025'
		 to view the email that the go server has sent.
	*/

	/* Here is how to do it (NOTES: How to send emails | create a channel |add a channel to a strruct)
	-There are a few go mailing packages, a good one of which is 'go-simple-mail'
	 (https://github.com/xhit/go-simple-mail). Run this cmd to install it:

	 	go get github.com/xhit/go-simple-mail/v2

	-To send an email, we have to know certain things:
		-the to address
		-the from address
		-the subject
		-the msg content

	-You go into your models and define all of that data in a new struct called eg MailData.
	 It being in a model means its imported into different parts of your app from where you
	 will be be able to easily send off emials.

			// MailData holds an email message
			type MailData struct {
				To string
				From string
				Subject string
				Content string
			}

	-We could easily create a function which will use whenerver we are sending emails & just pass
	 in the required params, but we will doing it in an even cool way; using channels so parts of your
	 app can listen in on that channel for sent emails. We will create a channel and make it available
	 all over your app. That channel will serve one purpose only-it will listen for a data type of
	 models.MaildData (the struct we just created in '/internal/models/model.go').
	-The question is where do we put it. If it is to be aailable to al parts of your app, it's best
	 to put it in the App.Config struct in /internal.config.config.go that is already being passed
	 around & made available throughout your app. Therefore, in your config.go add an entry to the
	 struct and its type is going to to be a channel:

	 		type AppConfig struct {
				...
				MailChan		chan models.MailData
			}

	-Next we have to actually create the channel. We can create in our main.go file:

			mailChan := make(chan models.MailData)
			//add it to the app's config (struct)
			app.MailChan = mailChan

	-Next we need a place in our app to listen on this channel. Create a new file eg called
	 send-mail.go in your app eg in /web/send-mail.go. Inside this file write the listening
	 code:

			package main

			import (
				"log"
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

	/*
				client, err := server.Connect()
				if err != nil {
					errorLog.Println(err)
				}

				//Construct your email message in a format that our client understands. We do this using m
				// mail has a method NewMSG() below which returns a struct which is a pointer to email & a new email msg
				email := mail.NewMSG()
				email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
				//email.SetBody(mail.TextHTML, "Hello, <strong>world</strong>!")
				email.SetBody(mail.TextHTML, m.Content)

				err = email.Send(client)
				if err != nil {
					errorLog.Println(err)
				} else {
					log.Println("Email sent!")
				}
			}


		-Next, we just have to make sure that the listenForMail() is actually called. You can
		 set that up in main.go ideally right after you're deferring a close on the mail channel

		 		defer close(app.MailChan)
				listenForMail()

		-Finally, create & send an email message. I would do it in main.go right after
			calling listenForMail() eg:

				msg := models.MailData{
					To:      "john@do.ca",
					From:    "me@here.ca",
					Subject: "Some subject",
					Content: "",
				}

				app.MailChan <- msg

		-To test emails, its best to do so using a dummy testing email server. There are various
		 options eg mailtrap.io, and MailHog. Let's use MailHog.
		-To install MailHog on a Mac
		 	-visit: https://github.com/mailhog/Mailhog
			-Scroll down to the installation text. Among the options are installing via Homebrew.
			 However you will need to have Homebrew first. Install it like so:

			 	/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install.sh)"

			Homebrew will take a while to install, but once its done, you can then use it to install MailHog
			like so:

				brew update && brew install mailhog

			Once mailhog is installed, you can start it by running any of these commands in the command line,
			ideally in a new CLI tab:

					mailhog
					OR
					brew services start mailhog

				Stop it with this command:

					bew services stop mailhog

			-View your emails sent from your Go application by visiting:
					http://localhost:8025

		-To install MailHog on Windows, again visit the site: https://github.com/mailhog/Mailhog
			Scroll down and click on the link to download it for Windows. Move the downloaded executable
			file to your desktop then double-click on it to start it up. Once its up and running,
			go to your browser and visit this link and port on your localhost:
				http://localhost:8025
			and you should see MailHog.

			Leave the emial server running while you test the sending of emails in your application.

		-Now let's create an email. We intend to send two emails after a reservation is made;
		 one to the owner of the property where the reservation was made, and
		 one to the user that made the reservation. It makes sense therefore to sned the emails in
		 the PostReservation() handler right after we have validated the request, and inserted the
		 reservation and the restriction, just before we redirect the user to the home page.

			//-------------------------------------------
			// send email notifications - first to guest
			htmlMessage := fmt.Sprintf(`
					<strong>Reservation Confirmation</strong><br>
					Dear %s, <br>
					This is to confirm your reservation from %s to %s.
				`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"),
				reservation.EndDate.Format("2006-01-02"))

			msg := models.MailData{
				To:      reservation.Email,
				From:    "gustavfn@yahoo.co.uk",
				Subject: "Reservation Confirmation",
				Content: htmlMessage,
			}

			m.App.MailChan <- msg
			//-------------------------------------------

			//-------------------------------------------
			// send email notifications - to property owner
			htmlMessage = fmt.Sprintf(`
					<strong>Reservation Notification</strong><br>
					Dear %s, <br>
					This is to notify you of a new reservation that has been booked for your property%s.<br>
					The Booking is by %s %s and the reservation is <br>
					from %s to %s.<br>

					Kind regards<br>
					The dream team
				`, "IDoNotKnowOwnerName", reservation.Room.RoomName, reservation.FirstName, reservation.LastName,
				reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

			msg = models.MailData{
				To:      "IDoNotKnowOwnerEmail@gmail.com",
				From:    "gustavfn@yahoo.co.uk",
				Subject: "Reservation Notification",
				Content: htmlMessage,
			}

			m.App.MailChan <- msg
			//-------------------------------------------


		-Now test this by restarting your application server. Keep the MailHog email server
		 running. Visit your application and make a reservation. Once the reservation has been
		 completed successfully, check for the new email to yourself as the guest at the MailHog
		 host page:  http://localhost:8025 and you should see that the new email has arrived.

		-Making HTML emails look prettier with Foundation the CSS framework
			visit: https://get.foundation/emails.html
			-Click on Get started > choose the non-sass version and go for the CSS omly one. So click
			 on 'Download Starter Kit' to download the CSS starter kit
			-Click to Download Foundation for Emails
			-extract the zip file, you will see a CSS and a templates folder, and an index.html file
			-Open the templates folder and you can preview them. Choose the eone you like eg drip.html
			-open the one you choose in your text editor. It contains all the code (CSS & HTML)
			-Copy all the code, create a new directory eg 'email-templates' in your app's root
			-Create a file in there eg 'basic.html' & paste the code in there.
			-Go into the code & remove the bits you do not need eg buttons, text, images, etc and just
			  customise it to your app. Basically, go in there, find the main p tag with lots of text,
			  remove the text and add a placeholder there like [%body]. You can have many of these
			  placeholders in different places on the page. The way you would use this is, before
			  sending an email, you would grab this email template content, substitute all the
			  placeholders with data for the email, and send off the email using that template as the
			  message content.

		-Go into your internal/models/model.go file and modify the MailData by adding a Template

			type MailData struct {
				...
				Template string
			}

			The idea is to program the use of email templates when sending an email such that
			if a template is specified, it will be used, else it will be sent without the use
			of a template as we did during the testing.
		-With that said, go into the 'cmd/web/send-mail.go' file and replace this line:

				email.SetBody(mail.TextHTML, m.Content)

			With:

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

		-Next, where we are sending the emails (in our case in internal/handlers/handlers.go), modify the
			MailData struct property by adding the Template value, which should be a string of the template
			filename eg:

					msg := models.MailData{
						...
						Template: "basic.html",
					}

			Do this for both case when sending an email to the guest as well as to the property owner.
			Remember that if you do not need a template, you just have to set that Template value to a
			blank string or leave it out all together as its default value is a blank string.
		-Restart your go application and test the emails again.



	*/
	//----------------------end sending email with standard library------------------------

	fmt.Printf(fmt.Sprintf("Starting application on port %s", portNumber))

	serve := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = serve.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func run() (*driver.DB, error) {
	// Register the models.Reservation type with gob
	// What kind of stuff will i be putting in the session. Register them all here
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	// read flags
	inProduction := flag.Bool("production", true, "Application is in production")
	useCache := flag.Bool("cache", true, "Use template cache")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPass := flag.String("dbpass", "", "Database password")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

	flag.Parse()

	if *dbName == "" || *dbUser == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}
	/*
		NOTES: The above 'read flags' section is how you create commands to be used in the CLI.
		You use the built-in flag object
		when it comes to postgres the values for using SSL are
			disable (default)
			prefer (use an SSL certificate if it exists)
			require (don't start unless you have an SSL certificate on the server)
		-You have to declare data type of the value you will be assigning to this flag variable by
		 calling the relevant type casting function on the flag object like eg flag.Bool() for
		 booleans, or flag.String() for strings etc

		These variables will now be created as pointers, since that's how they are used in the CLI.

		To use them, if there is a variable in your code eg your main.go file where you have a variable
		whose value you want to set dynamically, for example a variable named 'inProduction', just assign
		its value to the inProduction read flag variable above like so:

			app.InProduction = *inProduction
			app.UseCache	 = *useCache

		That way, when you are ready to launch your app in production, after moving your code to the
		production server, you will simply run the command created by the flag to set the value of
		inProduction to true (and set the values of any other variables you have defined this same way).
		In simple words, the ideal use case for these flag variables is for any variables that will have
		a different value based on the environment you are in. This way, you then have commands that
		you can run in the separate environments to set the value of those variables for that specific
		environment. Therefore, typical candidates for these are:
			-variables that determine if we are in development or producttion
			-variables that determine whether to use cache or not
			-database connection credentials

		To use these read flag variables or command line flags as they are often referred to, to create
		a connection to your DB, do it like so:

			connectionString :=
				fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
					*dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)
			db, err := driver.ConnectSQL(connectionString)

		-Notice how you have to call Parse() on the flag object after defining all the flag variables
		 for them to take effect before you can use them.

			flag.Parse()

		-The if statemant after the flag.Parse() is how you validate, and throw a CLI message to inform
		 the user if they failed to pass a required option in the command they ran.

		 	if *dbName == "" || *dbUser == "" {
				fmt.Println("Missing required flags")
				os.Exit(1)
			}

		-Having set all that up, you can now define the command based on thos variables in your
		 run.sh file that lives in the root of your application like so:

		 	./bookings -dbname=hotel-bookings -dbuser=user -cache=false production=false

		  Where:
		  	the value of dbname (in our case above 'hotel-bookings') should be the actual name of
				your application's database.
		  	the value of dbuser (in our case above 'user') should be the actual user
		  of the DB host your application is going to try to connect to.

		  If you just try to run the application (in the CLI like so: ./run.sh) without this line
		  in your run.sh file:

		 	./bookings -dbname=hotel-bookings -dbuser=user -cache=false production=false

		  which means you are trying to run the application without any of the required flags
		  (in our case defined above dbName or dbUser), you will get the CLI error message you
		  defined above in the print command:

		  	"Missing required flags"

		  If you put back this like in your run.sh file,

		 	./bookings -dbname=hotel-bookings -dbuser=user -cache=false production=false

		  then run the application again in the CLI like so:

		  	./run.sh

		  it should run just fine. Test that everything works as expected, ie if the read flag
		  variables have the expected values. Try refreshing your app in the browser to see if
		  it works fine. Try login in too.
	*/

	// NOTES: How to create a channel
	//------------------------------------
	mailChan := make(chan models.MailData)
	//add it to the app's config (struct)
	app.MailChan = mailChan
	// Note that we close this channel coz all channels must be closed for performance reasons
	// We cannot close it in this run() func since its like an init func that runs once the app
	// starts & sets up everything, if we clode it here it will close the channel as soon as it
	// creates it-hence we close it outside this channel, at the start of this main() func like
	// so: defer close(app.MailChan) in the same way we close the DB connect func (defer db.SQL.Close())
	//------------------------------------

	// change this to true when in production
	app.InProduction = *inProduction
	app.UseCache = *useCache

	// set up logging. Create a new logger that writes to the terminal (os.Stdout), prefix the msg
	// with "INFO" & a tab, followed by the date & time
	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// initialise a session
	session = scs.New()

	// optionally set lifetime of session
	// 24 hours. A syntax error in this time specification will cause the session setting & retrieving of data not to work
	session.Lifetime = 24 * time.Hour

	// Name sets the name of the session cookie. It should not contain
	// The default cookie name is "session".
	// If your application uses two different sessions, you must make sure that
	// the cookie name for each of these sessions is unique.
	session.Cookie.Name = "testProj_session_id"
	//by default it uses cookie for itas data storage, but it has different storages u can choose from eg DBs
	session.Cookie.Persist = true // should the cookie persist after user closes the browser?
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction // set to true when using https in production

	app.Session = session

	// connect to DB
	log.Println("Connecting to DB")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)
	db, err := driver.ConnectSQL(connectionString)
	// connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s" port=5432 hotel-bookings user=user ")
	//db, err := driver.ConnectSQL("host=localhost port=5432 dbname=hotel-bookings user=user password=")
	if err != nil {
		log.Fatal("Cannot cronnecting to database! Dying...")
	}
	log.Println("Connected to database")

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache")
		return nil, err
	}

	app.DefaultAppTitle = "Hotel Reservation App"
	app.TemplateCache = templateCache

	//set things up with our handlers
	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
