package emailer

import (
	"fmt"
	"github.com/mikedonnici/rtcl-api/datastore"
	"os"
)

// WelcomeUser sends a welcome email with a link to unlock the user account.
func WelcomeUser(u datastore.User) {

	body := `<h3>Welcome, %s!</h3>
             <p>Please click on this link to activate your account:</p>
			 <p><a href="%s" target="_blank">Activate my account</a></p>
			 <p>Happy RTCL-ing</p>`
	link := os.Getenv("API_URL") + "/users/" + u.ID.Hex() + "/confirm/" + u.KeyGen()
	body = fmt.Sprintf(body, u.FirstName, link)

	e := New()
	e.FromEmail = "notifier@rtcl.io"
	e.FromName = "RTCL Notifier"
	e.Subject = "Welcome to RTCL"
	e.ToEmail = u.Email
	e.ToName = u.FirstName + " " + u.LastName
	e.PlainContent = "Confirmation link: " + link
	e.HTMLContent = body
	e.Send()
}

// ResetPassword sends an email with a link to reset the user password.
func ResetPassword(u datastore.User) {

	body := `<h3>Hi, %s!</h3>
			 <p>The link below will allow you to reset your password.</p>
             <p>If you didn't ask for this you can ignore this email.</p>
			 <p><a href="%s" target="_blank">Reset my password</a></p>
			 <p>Happy RTCL-ing</p>`
	link := os.Getenv("API_URL") + "/users/" + u.ID.Hex() + "/reset/" + u.KeyGen()
	body = fmt.Sprintf(body, u.FirstName, link)

	e := New()
	e.FromEmail = "notifier@rtcl.io"
	e.FromName = "RTCL Notifier"
	e.Subject = "RTCL password reset"
	e.ToEmail = u.Email
	e.ToName = u.FirstName + " " + u.LastName
	e.PlainContent = "Password reset link: " + link
	e.HTMLContent = body
	e.Send()
}
