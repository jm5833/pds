/*
	Jack Martinez
	Semester Project - Part 4
	N18522789
	Application Server
	Expected run format: go run backend.go [file directory] [port number]
	[file directory] - directory where the 3 shared files are
					   going to be stored in
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
	"sync"
//	"runtime"
)

//global locks for each file
var user_lock = &sync.Mutex{}
var post_lock = &sync.Mutex{}
var follow_lock = &sync.Mutex{}
//global lock used for the condition variable
var message_lock = &sync.Mutex{}
var condVar *sync.Cond
//string to pass messages to the replication server
var message string
//filename consts
var USER_FILENAME string = "user.txt"
var POST_FILENAME string = "post.txt"
var FOLLOW_FILENAME string = "follow.txt"

func main(){
	//get the port number for the backend to run on
	if len(os.Args) < 3{
		log.Fatal("Not enough arguments. Need the directory and port number\n")
	}
	setupDirectory(os.Args[1])
	port := ":" + os.Args[2]
	server,serr := net.Listen("tcp", port)
	checkError(serr, 1)
	go replicationServer()
	acceptClient(server)
}
//function that maintains communication with the replication server
func replicationServer(){
	replication,err := net.Dial("tcp", ":8999")
	checkError(err, 1)
	go handleReplicationMessage(replication)
	condVar = sync.NewCond(message_lock)
	for true{
		message_lock.Lock()
		for len(message) == 0{
			condVar.Wait()
		}
		fmt.Println("Message to RS: " + message)
		replication.Write([]byte(message + "\n"))
		message = ""
		message_lock.Unlock()
	}
}
//function to handle the input from the replication server
func handleReplicationMessage(replication net.Conn){
	repbuf := bufio.NewScanner(replication)
	for repbuf.Scan(){
		retval := repbuf.Text()
		performAction(retval, replication)
	}
}
//function to setup the directory for where the files will be stored
func setupDirectory(directory string){
	if directory[:len(directory) - 1] != "/"{
		directory = directory + "/"
	}
	USER_FILENAME = directory + USER_FILENAME
	POST_FILENAME = directory + POST_FILENAME
	FOLLOW_FILENAME = directory + FOLLOW_FILENAME
}
//function to accept clients
func acceptClient(server net.Listener){
	for{
		client, err := server.Accept()
		checkError(err, 1)
		fmt.Println("Client has connected")
		go handleClient(client)
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
		sendToReplication(data)
	}
}
//function to send the data to the replication thread
func sendToReplication(data string){
	index := strings.Index(data, ";")
	if index == -1{
		return
	}
	action := data[:index]
	message_lock.Lock()
	switch action{
	case "Register":
		message = data
	case "Post":
		message = data
	case "Follow":
		message = data
	case "Unfollow":
		message = data
	case "Delete":
		message = data
	case "Default":
		message_lock.Unlock()
		return
	}
	message_lock.Unlock()
	condVar.Broadcast()
}
//function to perform what is in the action string
func performAction(data string, client net.Conn){
	index := strings.Index(data, ";")
	if index == -1{
		return
	}
	action := data[:index]
	fmt.Println(data)
	switch action{
	case "Register":
		register(data,client)
	case "Login":
		login(data,client)
	case "Post":
		post(data,client)
	case "GetPost":
		getPost(data,client)
	case "Follow":
		follow(data,client)
	case "Unfollow":
		unfollow(data,client)
	case "Delete":
		deleteProfile(data,client)
	case "GetFollowed":
		getFollowed(data,client)
	case "GetNotFollowed":
		getNotFollowed(data,client)
	}
}
//function to check the argument length
//returns a split of arguments if the correct length is given
//returns 1 if it has the correct argument length, 0 otherwise 
func checkArguments(data string, expected int, client net.Conn) ([]string,bool){
	//split the data
	args := strings.Split(data, ";")
	//check the length against the expected length
	if len(args) == expected{
		return args, true
	}
	client.Write([]byte("BadAL\n"))
	return args,false
}
//function to do error checking
//either exits or continues based on the checkType variable
func checkError(reterr error, checkType int){
	switch checkType{
	case 0:
		if reterr != nil{
			log.Println(reterr)
		}
	case 1:
		if reterr != nil{
			log.Fatal(reterr)
		}
	}
}
//function to register the user
//expected message format Register;[username];[password]
func register(data string, client net.Conn){
	//lock the mutex used for the user file
	user_lock.Lock()
	follow_lock.Lock()
	defer follow_lock.Unlock()
	defer user_lock.Unlock()
	//create and check the arguments passed in
	args, check := checkArguments(data, 3, client)
	if !check{
		return
	}

	//open/create the file that stores the username/password combinations
	userFile,deserr := os.OpenFile(USER_FILENAME, os.O_APPEND | os.O_CREATE | os.O_WRONLY,0664)
	checkError(deserr,1)
	defer userFile.Close()
	//read the contents of the user.txt file
	fileContent, filerr := ioutil.ReadFile(USER_FILENAME)
	checkError(filerr, 1)
	//check that the username/password isn't blank
	if len(args[1]) == 0 || len(args[2]) == 0{
		client.Write([]byte("BadAL\n"))
		return
	}else if strings.Contains(string(fileContent), args[1]){
		client.Write([]byte("UserExists\n"))
	}else{
		//Write to the user file in the format [Username],[Password]
		_,werr := userFile.WriteString(args[1] + "," + args[2] + "\n")
		checkError(werr, 0)
		client.Write([]byte("NewUser\n"))
		followFile, ferr := os.OpenFile(FOLLOW_FILENAME,
										os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
		_,werr = followFile.WriteString(args[1] + ",\n")
		checkError(ferr, 0)
	}
}
//function to log the user in
//expected message format: Login;[username];[password]
func login(data string, client net.Conn){
	//lock the user file
	user_lock.Lock()
	defer user_lock.Unlock()

	args, check := checkArguments(data, 3, client)
	if !check{
		return
	}
	//read the contents of the file
	fileContent, filerr := ioutil.ReadFile(USER_FILENAME)
	checkError(filerr, 0)
	//check for blank arguments
	if len(args[1]) == 0 || len(args[2]) == 0{
		client.Write([]byte("BadAL\n"))
		return
	}else if strings.Contains(string(fileContent), args[1] + "," + args[2]){
		client.Write([]byte("LoginOK\n"))
	}else{
		client.Write([]byte("BadAL\n"))
	}
}
//add a post
//expected message format Post;[username];[content];[timestamp]
func post(data string, client net.Conn){
	//get the lock for the post file
	post_lock.Lock()
	defer post_lock.Unlock()

	args, check := checkArguments(data, 4, client)
	if !check{
		return
	}

	postFile,deserr := os.OpenFile(POST_FILENAME, os.O_APPEND | os.O_CREATE | os.O_WRONLY,0664)
	checkError(deserr, 1)
	defer postFile.Close()

	//expecting the format of Post;[username];[content];[timestamp]
	if len(args[1]) == 0 || len(args[2]) == 0 || len(args[3]) == 0{
		client.Write([]byte("BadAL\n"))
	}
	client.Write([]byte("PostOK\n"))
	//write the user's post to a file
	_,writeerr := postFile.WriteString(args[1] + "," + args[2] + "," + args[3] + "\n")
	checkError(writeerr, 0)
}
//get a list of posts that the user follows
//expected message format GetPost;[username]
func getPost(data string, client net.Conn){
	//get the lock for the post and follow files
	post_lock.Lock()
	follow_lock.Lock()
	defer follow_lock.Unlock()
	defer post_lock.Unlock()

	args, check := checkArguments(data, 2, client)
	if !check{
		client.Write([]byte("End\n"))
		return
	}
	followContent, ferr := ioutil.ReadFile(FOLLOW_FILENAME)
	if ferr != nil{
		log.Println(ferr)
		client.Write([]byte("End\n"))
		return
	}
	postContent, perr := ioutil.ReadFile(POST_FILENAME)
	if perr != nil{
		log.Println(perr)
		client.Write([]byte("End\n"))
		return
	}

	followFile := strings.Split(string(followContent), "\n")
	postFile := strings.Split(string(postContent), "\n")

	//index that points to who the user follows
	followIndex := -1
	//figure out who the user follows
	for i, line := range followFile{
		userIndex := strings.Index(string(line), ",")
		if userIndex != -1 && line[:userIndex] == args[1]{
			followIndex = i
			break;
		}
	}
	if followIndex == -1{
		return
	}
	//send the post to the client if poster is on the user's follow list
	for i, line := range postFile{
		//users we will display posts from
		followList := strings.Split(followFile[followIndex], ",")
		post := strings.Split(string(line), ",")
		for _,user := range followList{
			if user == string(post[0]){
				client.Write([]byte(postFile[i] + "\n"))
			}
		}
	}
	client.Write([]byte("End\n"))
}
//delete the user profiFe
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
	//get the user file lock
	user_lock.Lock()
	defer user_lock.Unlock()

	userContent, uerr := ioutil.ReadFile(USER_FILENAME)
	checkError(uerr, 1)

	userFile,derr := os.OpenFile(USER_FILENAME, os.O_RDWR,0664)
	checkError(derr, 1)
	defer userFile.Close()

	userList := strings.Split(string(userContent), "\n")
	//array to hold the new user file
	var newUserFile []string
	//create a copy of the file while excluding the user we're deleting
	for _,line := range userList{
		index := strings.Index(line, ",")
		//check if the user we're checking is the user we want to delete
		if index == -1 || line[:index] == user{
			continue
		}
		newUserFile = append(newUserFile, line)
	}
	//truncate the file so that our writes happen at the beginning of the file
	terr := userFile.Truncate(0)
	checkError(terr, 1)
	//write the new contents of the file
	for _,line := range newUserFile{
		_,werr := userFile.WriteString(line + "\n")
		checkError(werr, 1)
	}
}
//delete all posts made by the user
func removeFromPost(user string){
	//get the post file lock
	post_lock.Lock()
	defer post_lock.Unlock()

	postContent, perr := ioutil.ReadFile(POST_FILENAME)
	if perr != nil{
		log.Println(perr)
		return
	}

	postFile,derr := os.OpenFile(POST_FILENAME, os.O_RDWR,0644)
	if derr != nil{
		log.Println(derr)
		return
	}
	defer postFile.Close()

	postlist := strings.Split(string(postContent), "\n")
	var newPostFile []string
	for _,line := range postlist{
		index := strings.Index(line, ",")
		if index == -1 || line[:index] == user{
			continue;
		}
		newPostFile = append(newPostFile, line)
	}
	terr := postFile.Truncate(0)
	checkError(terr, 1)

	for _,line := range newPostFile{
		_,werr := postFile.WriteString(line + "\n")
		checkError(werr, 1)
	}
}
//delete the user from everyone's follow list 
func removeFromFollow(user string){
	//get the follow file lock
	follow_lock.Lock()
	defer follow_lock.Unlock()

	followContent, ferr := ioutil.ReadFile(FOLLOW_FILENAME)
	checkError(ferr, 1)

	followFile,derr := os.OpenFile(FOLLOW_FILENAME, os.O_RDWR,0644)
	checkError(derr, 1)
	defer followFile.Close()

	followList := strings.Split(string(followContent), "\n")
	var newFollowFile []string

	for _,line := range followList{
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
			newFollowFile = append(newFollowFile, newline)
			continue
		}
		newFollowFile = append(newFollowFile, line)
	}
	terr := followFile.Truncate(0)
	checkError(terr, 1)

	for _,line := range newFollowFile{
		_,werr := followFile.WriteString(line + "\n")
		checkError(werr, 1)
	}
}
//function to add users to the follow list
//expected format: Follow;[user];[user to follow];...
func follow(data string, client net.Conn){
	//get the follow lock
	follow_lock.Lock()
	defer follow_lock.Unlock()
	args := strings.Split(data, ";")
	if len(args) < 3 || len(args[1]) == 0{
		client.Write([]byte("BadAL\n"))
		return
	}
	followFile, derr := os.OpenFile(FOLLOW_FILENAME, os.O_APPEND | os.O_CREATE | os.O_WRONLY,0644)
	checkError(derr, 0)
	defer followFile.Close()

	followContent, ferr := ioutil.ReadFile(FOLLOW_FILENAME)
	checkError(ferr, 1)

	followList := strings.Split(string(followContent), "\n")
	var newFollowList []string
	index := -1

	for i,line := range followList{
		j := strings.Index(line, ",")
		if j != -1 && line[:j] == args[1]{
			index = i
		}
		newFollowList = append(newFollowList, line)
	}
	//add a new entry if the user doesn't have a follow list yet
	if index == -1{
		//skip the first argument because its the follow command
		for i := 1; i < len(args); i++{
			followFile.WriteString(args[i] + ",")
		}
		followFile.WriteString("\n")
	}else if index != -1{
		//start at 2 because we're skipping Follow and [username]
		for i := 2; i < len(args); i++{
			if len(args[i]) == 0{
				continue
			}
			//check to make sure we don't add the same user twice
			if !strings.Contains(newFollowList[index], args[i]){
				newFollowList[index] = newFollowList[index] + args[i] + ","
			}
		}
		terr := followFile.Truncate(0)
		checkError(terr, 0)
		for _,line := range newFollowList{
			_,werr := followFile.WriteString(line + "\n")
			checkError(werr, 1)
		}
	}
	client.Write([]byte("FollowOK\n"))
}
//function to unfollow users
//expected format: Unfollow;[username];[user];...
func unfollow(data string, client net.Conn){
	//get the follow lock
	follow_lock.Lock()
	defer follow_lock.Unlock()

	args := strings.Split(data, ";")
	if len(args) < 3 || len(args[1]) == 0{
		client.Write([]byte("BadAL\n"))
		return
	}
	followFile, derr := os.OpenFile(FOLLOW_FILENAME, os.O_RDWR, 0644)
	checkError(derr, 1)
	defer followFile.Close()

	followContent, ferr := ioutil.ReadFile(FOLLOW_FILENAME)
	checkError(ferr, 0)

	followList := strings.Split(string(followContent), "\n")
	var newFollowList []string
	index := -1
	for i,line := range followList{
		j := strings.Index(line, ",")
		if j != -1 && line[:j] == args[1]{
			index = i
		}else if len(line) == 0 || j == -1{
			continue
		}
		newFollowList = append(newFollowList, line)
	}
	if index == -1{
		log.Fatal("bad unfollow list\n")
	}else{
		for i := 2; i < len(args); i++{
			if len(args[i]) == 0{
				continue
			}
			newFollowList[index] = strings.Replace(newFollowList[index], "," + args[i], "", -1)
		}
		followFile.Truncate(0)
		for _,line := range newFollowList{
			followFile.WriteString(line + "\n")
		}
	}
	client.Write([]byte("UnfollowOK\n"))
}
//get a list of all followed users
//expected message format: GetFollowed;[username]
func getFollowed(data string, client net.Conn){
	//get the follow file lock
	follow_lock.Lock()
	defer follow_lock.Unlock()

	followContent,ferr := ioutil.ReadFile(FOLLOW_FILENAME)
	checkError(ferr, 1)

	args := strings.Split(data,";")
	if len(args) < 2 || len(args[1]) == 0{
		client.Write([]byte("BadAL\n"))
	}

	userindex := -1
	followlist := strings.Split(string(followContent), "\n")
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
	client.Write([]byte("End\n"))
}
//get a list of users that the user isn't following
//expected message format: GetNotFollowed;[username]
func getNotFollowed(data string, client net.Conn){
	//get the user and follow locks
	user_lock.Lock()
	follow_lock.Lock()
	defer follow_lock.Unlock()
	defer user_lock.Unlock()

	followContent, ferr := ioutil.ReadFile(FOLLOW_FILENAME)
	checkError(ferr, 1)
	userContent, uerr := ioutil.ReadFile(USER_FILENAME)
	checkError(uerr, 1)

	args := strings.Split(data, ";")
	if len(args) < 2 || len(args[1]) == 0{
		client.Write([]byte("BadAL\n"))
		return
	}

	followlist := strings.Split(string(followContent), "\n")
	userlist := strings.Split(string(userContent), "\n")
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
	client.Write([]byte("End\n"))
}
