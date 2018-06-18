/*
	Jack Martinez
	Semester Project - Part 1
	N18522789

*/
package main

import(
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"log"
	"text/template"
)

type User struct{
	Username string
	Password string
	Follow []string	//other users this user follows
}

type Post struct{
	Timestamp time.Time	//time post was made
	Username string		//user who made the post
	Content string		//post the user made
}

const COOKIE_UNAME string = "cookie_uname"
const DEFLEN int = 0
const DEFCAP int = 10
var allusers []User
var allposts []Post

func main(){
	allusers = make([]User, DEFLEN, DEFCAP)
	allposts = make([]Post, DEFLEN, DEFCAP)

	fmt.Println("starting...")
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/post", post)
	http.HandleFunc("/", home)
	http.HandleFunc("/home", home)
	http.HandleFunc("/myprofile",profile)
	http.HandleFunc("/logout", logout)
	http.ListenAndServe(":8080", nil)
}
//function to load pages onto the screen
func loadpage(w http.ResponseWriter, r *http.Request){
	url := r.URL.Path[1:] + ".html"
	page,err := ioutil.ReadFile(url)
	if err != nil{ log.Println(err)
	}else{ fmt.Fprintf(w, string(page)) }

}
//registration for the site
func register(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodGet:
		//check if the user is already logged in
		cookie_uname := getcookie(r, COOKIE_UNAME)
		//redirect the user home is they're already logged in
		if cookie_uname != nil{
			http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
		//load the page if the user isn't logged in
		}else{ loadpage(w,r) }
	case http.MethodPost:
		r.ParseForm()
		//check if the username is already taken and reload the page if it is
		pos := searchuser(r.PostFormValue("uname"))
		if pos != -1{ loadpage(w,r)
		}else{
			adduser(r.PostFormValue("uname"), r.PostFormValue("pword"))
			createcookie(w, COOKIE_UNAME, r.PostFormValue("uname"))
			//redirect the user home after they logged in
			http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
		}
	}
}
//add a user to allusers
func adduser(uname, pword string){
	allusers = append(allusers, User{Username:uname, Password:pword})
	pos := searchuser(uname)
	allusers[pos].Follow = append(allusers[pos].Follow, uname)
}
//print all registered users
func printusers(){
	for i := 0; i < len(allusers); i++{
		fmt.Println(allusers[i].Username)
	}
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
//gets the cookie thats named cookiename
func getcookie(r *http.Request,cookiename string) *http.Cookie{
	cookie,err := r.Cookie(cookiename)
	if err != nil{ log.Println(err) }
	return cookie
}
//login function
func login(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodGet:
		//check if the user is already logged in
		cookie_uname := getcookie(r,COOKIE_UNAME)
		//redirect the user home if they're already logged in
		if cookie_uname != nil{
			http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
		//load the page 
		}else{ loadpage(w,r) }
	case http.MethodPost:
		r.ParseForm()
		//check if the username/password combination exists
		check := checkuser(r.PostFormValue("uname"), r.PostFormValue("pword"))
		if check == 1{
			//take the user to their homepage if they log in
			createcookie(w,COOKIE_UNAME,r.PostFormValue("uname"))
			http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
		//reload the page if the user doesn't enter the correct credentials
		}else{ loadpage(w,r) }
	}
}
//check if the username/password combination given exists
//returns 1 if it does, 0 otherwise
func checkuser(uname, pword string) int{
	for i := 0; i < len(allusers); i++{
		if allusers[i].Username == uname && allusers[i].Password == pword{
			return 1
		}
	}
	return 0
}
//create a post
func post(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodGet:
		//check if the user is logged in
		cookie_uname := getcookie(r,COOKIE_UNAME)
		//redirect the user to login if they're not logged in
		if cookie_uname == nil{
			http.Redirect(w,r,"/login",http.StatusTemporaryRedirect)
		//load the page if the user is logged in
		}else{ loadpage(w,r) }

	case http.MethodPost:
		r.ParseForm()
		//get the user cookie
		cookie_uname := getcookie(r,COOKIE_UNAME)
		content := r.PostFormValue("content")
		addcontent(cookie_uname.Value, content)
		http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
	}
}
//adds a tweet to the allposts slice
func addcontent(username, content string){
	post := Post{
		Timestamp:	time.Now(),
		Username:	username,
		Content:	content,
	}
	allposts = append(allposts, post)
}
//prints all posts onto the terminal
func printposts(){
	for i := len(allposts) -1; i >= 0; i--{
		fmt.Println(allposts[i].Content)
	}
}
//function to display the home page
func home(w http.ResponseWriter, r *http.Request){
	r.URL.Path = "/home"
	logged,pos := 0, -1
	//load the page
	loadpage(w,r)
	//get the usercookie
	cookie_uname := getcookie(r,COOKIE_UNAME)
	if cookie_uname != nil{
		logged = 1
		pos = searchuser(cookie_uname.Value)
		fmt.Fprintf(w,"<a href='logout'> Logout </a></br>")
		fmt.Fprintf(w,"<a href='post'> Post </a></br>")
		fmt.Fprintf(w,"<a href='myprofile'> My Profile </a></br>")
	}else{
		fmt.Fprintf(w,"<a href='login'> Login </a></br>")
		fmt.Fprintf(w,"<a href='register'> Register </a></br>")
	}
	//template to be filled
	const datatmpl = "</br>{{.Username}} Posted at {{.Timestamp}}:{{.Content}}"
	found := 0
	for i := len(allposts) -1; i >= 0; i--{
		//creation of the template
		found = 0
		if logged == 1{
			//check if the user follows the person who made each post
			for j := 0; j < len(allusers[pos].Follow); j++{
				if allusers[pos].Follow[j] == allposts[i].Username{
					found = 1
				}
			}
		}
		//don't post if the post doesn't come from someone on the follow list
		if found == 0{ continue }
		//create the template
		tmpl, err := template.New("DataTemplate").Parse(datatmpl)
		if err != nil{ log.Println(err) }
		//execute the template
		err = tmpl.Execute(w, allposts[i])
		if err != nil{ log.Println(err) }
	}
}
//serach to see if uname is a valid user
func searchuser(uname string) int{
	pos := -1
	for i := 0; i < len(allusers); i++{
		if uname == allusers[i].Username{
			pos = i
		}
	}
	return pos
}
//function to log the user out
func logout(w http.ResponseWriter, r *http.Request){
	cookie_uname := getcookie(r,COOKIE_UNAME)
	//redirect the user to the home page if they're not logged in
	if cookie_uname == nil{
		http.Redirect(w,r,"/home",http.StatusFound)
	}else{
		//set the expiration date of the cookie to a time that passed
		//so that it essentially "logs" the user out
		cookie_uname.Expires = time.Unix(0,0)
		http.SetCookie(w, cookie_uname)
		http.Redirect(w,r,"/home",http.StatusFound)
	}
}
//function for the myprofile page
func profile(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodGet:
		cookie_uname := getcookie(r,COOKIE_UNAME)
		//redirect the user if they're not logged in
		if cookie_uname == nil{
			fmt.Println("User not logged in")
			http.Redirect(w,r,"/login",http.StatusTemporaryRedirect)
		//load the page if they're logged in
		}else{
			loadpage(w,r)
			fmt.Fprintf(w,"<h1> Who do you wish to follow? </h1>")
			fmt.Fprintf(w,"<form method='post'>")
			fmt.Fprintf(w,"<input type='hidden' name='follow' value='follow'/>")
			selectfollow(w,r)
			fmt.Fprintf(w,"<h1> Who do you wish to unfollow? </h1>")
			fmt.Fprintf(w,"<form method='post'>")
			fmt.Fprintf(w,"<input type='hidden' name='block' value='block'/>")
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
		if(r.FormValue("delete") != ""){
			deleteprofile(w,r)
		}
	}
}
//function to let user select users to follow
func selectfollow(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "<fieldset>")
	cookie_uname := getcookie(r, COOKIE_UNAME)
	userpos := searchuser(cookie_uname.Value)

	const datatmpl =
"<input type='checkbox' name='uname' value='{{.Username}}'/>{{.Username}}</br>"
	for i := 0; i < len(allusers); i++{
		//getting the index of the user to be displayed
		pos := find(allusers[i].Username, allusers[userpos].Follow)
		//don't print the username if the
		//user logged in already follows them
		if pos != -1{ continue }
		//create the template
		tmpl, err := template.New("DataTemplate").Parse(datatmpl)
		if err != nil{ log.Println(err) }
		//execute the template
		err = tmpl.Execute(w, allusers[i])
		if err != nil{ log.Println(err) }
	}
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Submit\"/>")
	fmt.Fprintf(w, "</fieldset></form>")

}
//function to let user select users to follow
func selectunfollow(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "<fieldset>")
	cookie_uname := getcookie(r, COOKIE_UNAME)
	userpos := searchuser(cookie_uname.Value)
	//I intentionally start at 1 because the first user
	//in follow is always the user thats logged in. I dont
	//think users should be able to unfollow themselves 
	for i := 1; i < len(allusers[userpos].Follow); i++{
		output := ""
		//getting the index of the user to be displayed
		pos := find(allusers[i].Username, allusers[userpos].Follow)
		//don't print the username if the
		//user logged doesn't follows them
		if pos == -1{ continue }
		output = "<input type='checkbox' name='uname'value='" +
					allusers[userpos].Follow[i] +
					"'/>" + allusers[userpos].Follow[i] +
					"</br>"
		fmt.Fprintf(w,output)
	}
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Submit\"/>")
	fmt.Fprintf(w, "</fieldset></form>")

}

//add users to the follow list
func followusers(w http.ResponseWriter, r *http.Request){
	cookie_uname := getcookie(r,COOKIE_UNAME)
	pos := searchuser(cookie_uname.Value)
	for i := 0; i < len(r.Form["uname"]); i++{
		found := find(r.Form["uname"][i],allusers[pos].Follow)
		//dont add if the user is already in the list
		if found != -1{ continue }
		allusers[pos].Follow = append(allusers[pos].Follow, r.Form["uname"][i])
	}
}
//Remove users from the user's follow list
func unfollowusers(w http.ResponseWriter, r *http.Request){
	cookie_uname := getcookie(r,COOKIE_UNAME)
	pos := searchuser(cookie_uname.Value)
	for i := 0; i < len(r.Form["uname"]); i++{
		for j := 0; j < len(allusers[pos].Follow); j++{
			//get the index of the user thats going to be removed from Follow
			remove := find(r.Form["uname"][j],allusers[pos].Follow)
			//remove the user from the Follow list
			allusers[pos].Follow = append(allusers[pos].Follow[:remove],
										  allusers[pos].Follow[remove+1:]...)
		}
	}
}
//returns the index of value in slice
//returns -1 otherwise
func find(value string, slice []string) int{
	pos := -1
	for i := 0; i < len(slice); i++{
		if slice[i] == value{
			pos = i
		}
	}
	return pos
}
//deletes the users profile and all posts they made
func deleteprofile(w http.ResponseWriter, r *http.Request){
	cookie_uname := getcookie(r,COOKIE_UNAME)
	//get the index of the user account that is getting deleted
	pos := searchuser(cookie_uname.Value)
	//deleting the user from the all users list
	allusers = append(allusers[:pos], allusers[pos+1:]...)
	//delete the users from everyone's follow list
	for i := 0; i < len(allusers); i++{
		userindex := find(cookie_uname.Value, allusers[i].Follow)
		allusers[i].Follow = append(allusers[i].Follow[:userindex],
									allusers[i].Follow[userindex +1:]...)
	}
	//delete the post that the user made
	for i := 0; i < len(allposts); i++{
		if allposts[i].Username == cookie_uname.Value{
			allposts = append(allposts[:i], allposts[pos+1:]...)
			i--
		}
	}
	http.Redirect(w,r,"/logout",http.StatusTemporaryRedirect)
}
