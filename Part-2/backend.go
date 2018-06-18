/*
	Jack Martinez
	Semester Project - Part 2
	N18522789
	Application Server
*/
package main

import(
	"fmt"
	"bufio"
	"io/ioutil"
	"net"
	"strings"
	"log"
	"os"
)


//filename consts
const USER_FILENAME = "user.txt"
const POST_FILENAME = "post.txt"
const FOLLOW_FILENAME = "follow.txt"

func main(){
	server,err := net.Listen("tcp", ":9000")
	if err != nil{
		log.Fatal(err)
	}else{
		acceptClient(server)
	}
}

//function to accept clients
func acceptClient(server net.Listener){
	for{
		client, err := server.Accept()
		if err != nil{
			log.Fatal(err)
		}else{
			fmt.Println("Client has connected")
			go handleClient(client)
		}
	}
}
//function to handle client connections 
func handleClient(client net.Conn){
	//buffer to read messages from the client
	readbuf := bufio.NewScanner(client)
	//buffer to write messages to send to the client
	for readbuf.Scan(){
		data := readbuf.Text()
		performAction(data, client)
	}
}
//function to perform what is in the action string
func performAction(data string, client net.Conn){
	index := strings.Index(data, ";")
	if index == -1{
		return
	}
	fmt.Println(data)
	action := data[:index]
	if action == "Register"{
		register(data, client)
	}else if action == "Login"{
		login(data, client)
	}else if action == "Post"{
		post(data, client)
	}else if action == "GetPost"{
		getPost(data, client)
	}else if action == "Follow"{
		follow(data,client)
	}else if action == "Unfollow"{
		unfollow(data,client)
	}else if action == "Delete"{
		deleteProfile(data, client)
	}else if action == "GetFollowed"{
		getFollowed(data, client)
	}else if action == "GetNotFollowed"{
		getNotFollowed(data, client)
	}
}
//function to register the user
//expected message format Register;[username];[password]
func register(data string, client net.Conn){
	//split the data up to make it easier to work with
	args := strings.Split(data, ";")
	//check if the message received is of the correct length
	if len(args) < 3{
		client.Write([]byte("BadAL\n"))
		return;
	}
	//open/create the file that stores the username/password combinations
	file,deserr := os.OpenFile(USER_FILENAME, os.O_APPEND | os.O_CREATE | os.O_WRONLY,0664)
	if deserr != nil{ log.Fatal(deserr) }
	defer file.Close()
	//read the contents of the user.txt file
	fileContent, filerr := ioutil.ReadFile(USER_FILENAME)
	if filerr != nil{ log.Fatal(filerr) }
	//check that the username/password isn't blank
	if len(args[1]) == 0 || len(args[2]) == 0{
		client.Write([]byte("BadUP\n"))
		return
	}else if strings.Contains(string(fileContent), args[1]){
		client.Write([]byte("UserExists\n"))
	}else{
		_,writeerr := file.WriteString(args[1] + "," + args[2] + "\n")
		if writeerr != nil{ log.Println(writeerr) }
		client.Write([]byte("NewUser\n"))
		follow("Follow;" + args[1], client)
	}
}
//function to log the user in
//expected message format: Login;[username];[password]
func login(data string, client net.Conn){
	//split the data to make it easier to work with
	args := strings.Split(data,";")
	//check len of the arguments
	if len(args) < 3{
		client.Write([]byte("BadAL\n"))
		return
	}
	file,deserr := os.OpenFile(USER_FILENAME, os.O_APPEND | os.O_CREATE | os.O_WRONLY,0664)
	if deserr != nil{ log.Fatal(deserr) }
	defer file.Close()
	//read the contents of the file
	fileContent, filerr := ioutil.ReadFile(USER_FILENAME)
	if filerr != nil{ log.Println(filerr) }
	//check for blank arguments
	if len(args[1]) == 0 || len(args[2]) == 0{
		client.Write([]byte("BadUP\n"))
		return
	}else if strings.Contains(string(fileContent), args[1] + "," + args[2]){
		client.Write([]byte("LoginOK\n"))
	}else{
		client.Write([]byte("BadUP\n"))
	}
}
//add a post
//expected message format Post;[username];[content];[timestamp]
func post(data string, client net.Conn){
	args := strings.Split(data,";")
	if len(args) < 4{
		client.Write([]byte("BadAL\n"))
	}
	file,deserr := os.OpenFile(POST_FILENAME, os.O_APPEND | os.O_CREATE | os.O_WRONLY,0664)
	if deserr != nil{ log.Fatal(deserr) }
	defer file.Close()

	//expecting the format of Post;[username];[content];[timestamp]
	if len(args[1]) == 0 || len(args[2]) == 0 || len(args[3]) == 0{
		client.Write([]byte("BadUP\n"))
	}
	client.Write([]byte("PostOK\n"))
	//write the user's post to a file
	_,writeerr := file.WriteString(args[1] + "," + args[2] + "," + args[3] + "\n")
	if writeerr != nil{ log.Println(writeerr) }
}
//get a list of posts that the user follows
//expected message format GetPost;[username]
func getPost(data string, client net.Conn){
	args := strings.Split(data,";")
	if len(args) < 2 || len(args[1]) == 0{
		client.Write([]byte("BadAL\n"))
		return
	}
	follow, ferr := ioutil.ReadFile(FOLLOW_FILENAME)
	if ferr != nil{
		log.Println(ferr)
		client.Write([]byte("End\n"))
		return
	}
	post, perr := ioutil.ReadFile(POST_FILENAME)
	if perr != nil{
		log.Println(perr)
		client.Write([]byte("End\n"))
		return
	}

	followfile := strings.Split(string(follow), "\n")
	postfile := strings.Split(string(post), "\n")

	//index that points to who the user follows
	followindex := -1
	//figure out who the user follows
	for i, line := range followfile{
		userindex := strings.Index(string(line), ",")
		if line[:userindex] == args[1]{
			followindex = i
			break;
		}
	}
	//send the post to the client if poster is on the user's follow list
	for i, line := range postfile{
		//users we will display posts from
		followlist := strings.Split(followfile[followindex], ",")
		post := strings.Split(string(line), ",")
		for _,user := range followlist{
			if user == string(post[0]){
				client.Write([]byte(postfile[i] + "\n"))
			}
		}
	}
	client.Write([]byte("End\n"))
}
//delete the user profile
//expected format Delete;[username]
func deleteProfile(data string, client net.Conn){
	args := strings.Split(data, ";")
	if len(args) < 2 || len(args[1]) == 0{
		client.Write([]byte("BadAL\n"))
		return;
	}
	removeFromUser(args[1])
	removeFromPost(args[1])
	removeFromFollow(args[1])
}
//remove the user from the user file
func removeFromUser(user string){
	userf, uerr := ioutil.ReadFile(USER_FILENAME)
	if uerr != nil{ log.Fatal(uerr) }

	file,derr := os.OpenFile(USER_FILENAME, os.O_RDWR,0664)
	if derr != nil{ log.Fatal(derr) }
	defer file.Close()

	userlist := strings.Split(string(userf), "\n")
	//array to hold the new user file
	var newuserfile []string
	//create a copy of the file while excluding the user we're deleting
	for _,line := range userlist{
		index := strings.Index(line, ",")
		//check if the user we're checking is the user we want to delete
		if index == -1 || line[:index] == user{
			continue
		}
		newuserfile = append(newuserfile, line)
	}
	//truncate the file so that our writes happen at the beginning of the file
	terr := file.Truncate(0)
	if terr != nil{	log.Fatal(terr) }
	//write the new contents of the file
	for _,line := range newuserfile{
		_,werr := file.WriteString(line + "\n")
		if werr != nil{ log.Fatal(werr) }
	}
}
//delete all posts made by the user
func removeFromPost(user string){
	//postMutex.Lock()
	//defer postMutex.Unlock()

	postf, perr := ioutil.ReadFile(POST_FILENAME)
	if perr != nil{
		log.Println(perr)
		return
	}

	file,derr := os.OpenFile(POST_FILENAME, os.O_RDWR,0644)
	if derr != nil{
		log.Println(derr)
		return
	}
	defer file.Close()

	postlist := strings.Split(string(postf), "\n")
	var newpostfile []string
	for _,line := range postlist{
		index := strings.Index(line, ",")
		if index == -1 || line[:index] == user{
			continue;
		}
		newpostfile = append(newpostfile, line)
	}
	terr := file.Truncate(0)
	if terr != nil{ log.Fatal(terr) }

	for _,line := range newpostfile{
		_,werr := file.WriteString(line + "\n")
		if werr != nil{ log.Fatal(werr) }
	}
}
//delete the user from everyone's follow list
func removeFromFollow(user string){

	followf, ferr := ioutil.ReadFile(FOLLOW_FILENAME)
	if ferr != nil{ log.Fatal(ferr) }

	file,derr := os.OpenFile(FOLLOW_FILENAME, os.O_RDWR,0644)
	if derr != nil{ log.Fatal(derr) }
	defer file.Close()

	followlist := strings.Split(string(followf), "\n")
	var newfollowfile []string

	for _,line := range followlist{
		index := strings.Index(line, ",")
		if index == -1 || line[:index] == user{
			continue;
		}
		if strings.Contains(line, user){
			//remove the user from the follow list
			newline := strings.Replace(line,"," + user, "", -1)
			if strings.Index(newline, ",") == -1{
				newline = newline + ","
			}
			newfollowfile = append(newfollowfile, newline)
			continue
		}
		newfollowfile = append(newfollowfile, line)
	}
	terr := file.Truncate(0)
	if terr != nil{ log.Fatal(terr) }

	for _,line := range newfollowfile{
		_,werr := file.WriteString(line + "\n")
		if werr != nil{ log.Fatal(werr) }
	}
}
//function to add users to the follow list
//expected format: Follow;[user];[user to follow];...
func follow(data string, client net.Conn){
	args := strings.Split(data, ";")
	if len(args) < 2{
		client.Write([]byte("BadAL\n"))
	}
	file, derr := os.OpenFile(FOLLOW_FILENAME, os.O_APPEND | os.O_CREATE | os.O_WRONLY,0644)
	if derr != nil{ log.Println(derr) }
	defer file.Close()

	followf, ferr := ioutil.ReadFile(FOLLOW_FILENAME)
	if ferr != nil{ log.Fatal(ferr) }

	followlist := strings.Split(string(followf), "\n")
	var newfollowlist []string
	index := -1

	for i,line := range followlist{
		j := strings.Index(line, ",")
		if j != -1 && line[:j] == args[1]{
			index = i
		}
		newfollowlist = append(newfollowlist, line)
	}
	//add a new entry if the user doesn't have a follow list yet
	if index == -1{
		//skip the first argument because its the follow command
		for i := 1; i < len(args); i++{
			fmt.Print(args[i] + ",")
			file.WriteString(args[i] + ",")
		}
		file.WriteString("\n")
	}else if index != -1{
		for i := 2; i < len(args); i++{
			if len(args[i]) == 0{
				continue
			}
			newfollowlist[index] = newfollowlist[index] + args[i] + ","
		}
		terr := file.Truncate(0)
		if terr != nil{ log.Println(terr) }
		for _,line := range newfollowlist{
			file.WriteString(line + "\n")
		}
	}
}
//function to unfollow users
//expected format: Unfollow;[username];[user];...
func unfollow(data string, client net.Conn){
	args := strings.Split(data, ";")
	if len(args) < 3 || len(args[1]) == 0{
		client.Write([]byte("BadAL\n"))
	}
	file, derr := os.OpenFile(FOLLOW_FILENAME, os.O_RDWR, 0644)
	if derr != nil{ log.Fatal(derr) }
	defer file.Close()
	followf, ferr := ioutil.ReadFile(FOLLOW_FILENAME)
	if ferr != nil{ log.Println(ferr) }
	followlist := strings.Split(string(followf), "\n")
	var newfollowlist []string
	index := -1
	for i,line := range followlist{
		j := strings.Index(line, ",")
		if j != -1 && line[:j] == args[1]{
			index = i
		}else if len(line) == 0 || j == -1{
			continue
		}
		newfollowlist = append(newfollowlist, line)
	}
	if index == -1{
		log.Fatal("bad unfollow list\n")
	}else{
		for i := 2; i < len(args); i++{
			if len(args[i]) == 0{
				continue
			}
			newfollowlist[index] = strings.Replace(newfollowlist[index], "," + args[i], "", -1)
		}
		file.Truncate(0)
		for _,line := range newfollowlist{
			file.WriteString(line + "\n")
		}
	}
}
//get a list of all followed users
//expected message format: GetFollowed;[username]
func getFollowed(data string, client net.Conn){

	followf,ferr := ioutil.ReadFile(FOLLOW_FILENAME)
	if ferr != nil{ log.Fatal(ferr) }

	args := strings.Split(data,";")
	if len(args) < 2 || len(args[1]) == 0{
		client.Write([]byte("BadAL\n"))
	}

	userindex := -1
	followlist := strings.Split(string(followf), "\n")
	for i,line := range followlist{
		index := strings.Index(line,",")
		if index != -1 && line[:index] == args[1]{
			userindex = i
			break
		}
	}
	message := followlist[userindex] + "\n"
	if userindex != -1{
		client.Write([]byte(message))
	}
}
//get a list of users that the user isn't following
//expected message format: GetNotFollowed;[username]
func getNotFollowed(data string, client net.Conn){
	followf, ferr := ioutil.ReadFile(FOLLOW_FILENAME)
	if ferr != nil{ log.Fatal(ferr) }
	userf, uerr := ioutil.ReadFile(USER_FILENAME)
	if uerr != nil{ log.Fatal(uerr) }

	args := strings.Split(data, ";")
	if len(args) < 2 || len(args[1]) == 0{
		client.Write([]byte("BadAL\n"))
	}

	followlist := strings.Split(string(followf), "\n")
	userlist := strings.Split(string(userf), "\n")
	userindex := -1
	//figure out who the user follows
	for i,line := range followlist{
		index := strings.Index(line,",")
		if index != -1 && line[:index] == args[1]{
			userindex = i
			break
		}
	}
	userfollow := strings.Split(followlist[userindex], ",")
	found := false
	var notfollowed []string
	//add the users who the user doesn't 
	//follow to a list to send to the client
	for _,userline := range userlist{
		if len(userline) == 0{
			continue
		}
		index := strings.Index(userline,",")
		for _,followline := range userfollow{
			if userline[:index] == followline{
				found = false
				break
			}
			found = true
		}
		if found{
			notfollowed = append(notfollowed, userline[:index])
			found = false
		}
	}
	var message string
	for _,line := range notfollowed{
		message = message + line + ","
	}
	message = message + "\n"
	client.Write([]byte(message))
}
