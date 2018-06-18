/*
	Jack Martinez
	Semester Project - Part 2
	N18522789
	Frontend server
*/
package main

import(
	"fmt"
	"net"
	"net/http"
	"io/ioutil"
	//"os"
	"bufio"
	"time"
	"log"
	"strings"
	//"text/template"
)
const COOKIE_UNAME string = "cookie_uname"
var server net.Conn
var err error

func main(){
	//start the server
	fmt.Println("Starting...")
	//Connect to the application server thats hosted on port 8001
	server,err = net.Dial("tcp4", ":9000")
	//exit if the frontend can't establish a connection to the backend server
	if err != nil{
		fmt.Println("application server connection error")
	}
	http.HandleFunc("/", home)
	http.HandleFunc("/home", home)
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/post", post)
	http.HandleFunc("/profile", profile)
	http.HandleFunc("/logout", logout)
	http.ListenAndServe(":8000",nil)
}
//create a cookie with the name cookiename and value cookievalue
func createcookie(w http.ResponseWriter, cookiename, cookievalue string){
	cookie := http.Cookie{
		Name:		cookiename,
		Value:		cookievalue,
		Expires:	time.Now().Add(1 * time.Hour),
	}
	http.SetCookie(w, &cookie)
}
//get the user cookie
func getcookie(r *http.Request, cookiename string) *http.Cookie{
	cookie,err := r.Cookie(cookiename)
	if err != nil{ log.Println(err) }
	return cookie
}
//function to load the webpage
func loadpage(w http.ResponseWriter, r *http.Request){
	url := r.URL.Path[1:] + ".html"
	fmt.Println("Trying to load " + url)
	page,err := ioutil.ReadFile(url)
	if err != nil{ log.Println(err) }
	fmt.Fprintf(w,string(page))
}
//function to create an account
func register(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodGet:
		cookie := getcookie(r, COOKIE_UNAME)
		if cookie != nil{
			fmt.Println("logged in")
			http.Redirect(w,r,"/home", http.StatusTemporaryRedirect)
		}else{
			loadpage(w,r)
		}
	case http.MethodPost:
		r.ParseForm()
		//get the data from the form
		username := r.PostFormValue("uname")
		password := r.PostFormValue("pword")
		//construct the message to send to the client
		message := "Register;" + username + ";" + password + "\n"
		fmt.Println(message)
		server.Write([]byte(message))
		//create the read buffer to recieve the message from the server
		readbuf := bufio.NewScanner(server)
		readbuf.Scan()
		serverResponse := readbuf.Text()
		if serverResponse == "NewUser"{
			createcookie(w,COOKIE_UNAME,username)
			http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
		}else{
			loadpage(w,r)
		}
	}
}
//login function
func login(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodGet:
		cookie := getcookie(r, COOKIE_UNAME)
		if cookie != nil{
				http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
		}else{
			loadpage(w,r)
		}
	case http.MethodPost:
		r.ParseForm()
		//get the data from the forms 
		username := r.PostFormValue("uname")
		password := r.PostFormValue("pword")
		//construct the message to send to the server
		message := "Login;" + username + ";" + password + "\n"
		server.Write([]byte(message))
		readbuf := bufio.NewScanner(server)
		readbuf.Scan()
		//check the server's response
		if readbuf.Text() == "LoginOK"{
			createcookie(w,COOKIE_UNAME,username)
			http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
		}else{
			loadpage(w,r)
		}
	}
}
//create a post
func post(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodGet:
		//check if the user is logged in
		cookie := getcookie(r, COOKIE_UNAME)
		//redirect the user if the user is logged in
		if cookie == nil{
			http.Redirect(w,r,"/login",http.StatusTemporaryRedirect)
		}else{
			loadpage(w,r)
		}
	case http.MethodPost:
		r.ParseForm()
		cookie := getcookie(r,COOKIE_UNAME)
		content := r.PostFormValue("content")
		content = strings.Replace(content, "\n", " ", -1)
		ts := time.Now()
		message := "Post;"+cookie.Value+";"+content+";"+ts.String()+"\n"
		server.Write([]byte(message))
		readbuf := bufio.NewScanner(server)
		readbuf.Scan()
		serverResponse := readbuf.Text()
		if serverResponse == "BadUP"{
			loadpage(w,r)
		}else{
			http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
		}
	}
}
//home page
func home(w http.ResponseWriter, r *http.Request){
	r.URL.Path = "/home"
	logged := false
	//load the page
	loadpage(w,r)
	//get the cookie
	cookie := getcookie(r,COOKIE_UNAME)
	if cookie != nil{
		logged = true
		fmt.Fprintf(w,  "<a href='logout'> Logout </a></br>" +
						"<a href='post'> Post </a></br>" +
						"<a href='profile'> My Profile </a></br>")
	}else{
		fmt.Fprintf(w,  "<a href='login'> Login </a></br>" +
						"<a href='register'> Register </a></br>")
	}
	if logged{
		message := "GetPost;" + cookie.Value + "\n"
		server.Write([]byte(message))
		readbuf := bufio.NewScanner(server)
		for readbuf.Scan(){
			serverResponse := readbuf.Text()
			responseList := strings.Split(serverResponse, ",")
			if serverResponse == "End"{
				break
			}else if len(serverResponse) != 0{
				data := "</br>" + responseList[0] + " posted at " +
						responseList[2] + ":" + responseList[1]
				fmt.Fprintf(w,data)
			}
		}
	}
}
//function for my profile
func profile(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodGet:
		cookie := getcookie(r,COOKIE_UNAME)
		//redirect the user if they're not logged in
		if cookie == nil{
			http.Redirect(w,r,"/login",http.StatusTemporaryRedirect)
		}else{
			loadpage(w,r)
			fmt.Fprintf(w,"<h1> Who do you want to follow? </h1>" +
						  "<form method='post'>" +
						  "<input type='hidden' name='follow' value='follow'/>")
			selectfollow(w,r)
			fmt.Fprintf(w,"<h1> Who do you want to unfollow? </h1>" +
						  "<form method='post'>" +
						  "<input type='hidden' name='block' value='block'/>")
			selectunfollow(w,r)
	}
	case http.MethodPost:
		r.ParseForm()
		if len(r.Form["uname"]) > 0 && r.PostFormValue("follow") != ""{
			followusers(w,r)
			http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
		}
		if len(r.Form["uname"]) > 0 && r.PostFormValue("block") != ""{
			unfollowusers(w,r)
			http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
		}
		if r.FormValue("delete") != ""{
			deleteprofile(w,r)
			http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
		}
	}
}
//function to let the user select who to follow
func selectfollow(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w,"<fieldset>")
	cookie := getcookie(r,COOKIE_UNAME)
	message := "GetNotFollowed;" + cookie.Value + "\n"
	server.Write([]byte(message))
	readbuf := bufio.NewScanner(server)
	readbuf.Scan()
	serverResponse := readbuf.Text()
	notFollowedList := strings.Split(serverResponse, ",")
	for _,line := range notFollowedList{
		if line == cookie.Value || len(line) == 0{
			continue
		}
		output := "<input type='checkbox' name='uname' value='" +
				  line + "'/>" + line + "</br>"
		fmt.Fprintf(w,output)
	}
	fmt.Fprintf(w,"<input type='submit' value='Submit'>" +
				  "</fieldset></form>")
}
//function to send the server a list of who to follow
func followusers(w http.ResponseWriter, r *http.Request){
	cookie := getcookie(r,COOKIE_UNAME)
	message := "Follow;" + cookie.Value + ";"
	for i := 0; i < len(r.Form["uname"]); i++{
		message = message + r.Form["uname"][i] + ";"
	}
	message = message + "\n"
	server.Write([]byte(message))

}
//select users to unfollow
func selectunfollow(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w,"<fieldset>")
	cookie := getcookie(r,COOKIE_UNAME)
	message := "GetFollowed;" + cookie.Value + "\n"
	server.Write([]byte(message))
	readbuf := bufio.NewScanner(server)
	readbuf.Scan()
	serverResponse := readbuf.Text()
	followlist := strings.Split(serverResponse,",")
	for _,line := range followlist{
		if len(line) == 0 || line == cookie.Value{
			continue
		}
		output := "<input type='checkbox' name='uname' value='" +
				  line + "'/>" + line + "</br>"
		fmt.Fprintf(w,output)
	}
	fmt.Fprintf(w,"<input type='submit' value='Submit'>" +
				  "</fieldset></form>")
}
//unfollow selected users
func unfollowusers(w http.ResponseWriter, r *http.Request){
	cookie := getcookie(r,COOKIE_UNAME)
	message := "Unfollow;" + cookie.Value + ";"
	for _,line := range r.Form["uname"]{
		message = message + line + ";"
	}
	message = message + "\n"
	server.Write([]byte(message))
}
//logout
func logout(w http.ResponseWriter, r *http.Request){
	cookie := getcookie(r,COOKIE_UNAME)
	//redirect the user if they're logged in, else log them out
	if cookie == nil{
		http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
	}else{
		cookie.Expires = time.Unix(0,0)
		http.SetCookie(w, cookie)
		http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
	}
}
//delete profile
func deleteprofile(w http.ResponseWriter, r *http.Request){
	cookie := getcookie(r,COOKIE_UNAME)
	message := "Delete;" + cookie.Value + "\n"
	server.Write([]byte(message))
	http.Redirect(w,r,"/logout",http.StatusTemporaryRedirect)
}
