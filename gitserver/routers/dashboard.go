package routers

import (
	"html/template"
	"net/http"

	"github.com/0x0Dx/x/gitserver/utils"
)

func Dashboard(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/dashboard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Title": "Dashboard",
	})
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/user/signin.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Title":   "Sign In",
		"AppName": utils.Cfg.App.Name,
	})
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Handle signup form submission
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")

		_ = name
		_ = email
		_ = password

		http.Redirect(w, r, "/user/signin", http.StatusFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/user/signup.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Title":   "Sign Up",
		"AppName": utils.Cfg.App.Name,
	})
}
