/*
	Jack Martinez
	Semester Project - Part 3
	N18522789
	Replication Server
*/
package main

import(
	"fmt"
	"log"
	"strconv"
	"net"
	"bufio"
	//"sync"
	"runtime"
)

//default number of servers in all
const DEFAULT_SERVER_NUMBER = 3
//list of messages that the frontend is expecting
var SERVER_MESSAGES []string = []string{"NewUser", "UserExists", "BadAL",
										 "LoginOK", "PostOK", "End",
										 "PostOK", "FollowOK", "UnfollowOK"}
func main(){
	server, serr := net.Listen("tcp", ":9000")
	fmt.Println("Selection server started on localhost:9000")
	checkError(serr, 1)
	acceptClient(server)
}
//function to check error values passed in
func checkError(reterr error, checkType int){
	switch checkType{
	case 0:
		if reterr != nil{
			log.Println(reterr)
		}
	case 1:
		if reterr != nil{
			panic(reterr)
		}
	}
}
//function to accept the incoming client connection
func acceptClient(server net.Listener){
	for{
		client, cerr := server.Accept()
		checkError(cerr, 1)
		go handleClient(client)
	}
}
//function to handle the messages between server and client
func handleClient(client net.Conn){
	fmt.Println("Client connected")
	server := selectServer()
	serverChan := make(chan net.Conn)
	messageChan := make(chan string)
	go handleMessages(server, client, serverChan, messageChan)
	for{
		<-messageChan
		server = selectServer()
		serverChan <- server
	}
}
func handleMessages(server, client net.Conn, serverChan chan net.Conn, messageChan chan string){
	clientbuf := bufio.NewScanner(client)
	for{
		var messageList []string
		clientbuf.Scan()
		clientMessage := clientbuf.Text()
		fmt.Println("Client: " + clientMessage)
		mlen, err := server.Write([]byte(clientMessage + "\n"))
		//if the server dies, select a new server and resend the client message
		if mlen == 0 && err != nil{
			messageChan <- "hi"
			server = <-serverChan
			server.Write([]byte(clientMessage + "\n"))
		}
		//compile the list of messages that needs to be sent to the client
		serverbuf := bufio.NewScanner(server)
		for serverbuf.Scan(){
			serverResponse := serverbuf.Text()
			messageList = append(messageList, serverResponse)
			end := checkResponse(serverResponse)
			if end{
				fmt.Println("break")
				break
			}
		}
		for _,msg := range messageList{
			mlen, err := client.Write([]byte(msg + "\n"))
			fmt.Println("Server: " + msg)
			//exit if the client is dead
			if mlen == 0 && err != nil{
				fmt.Println("Exiting.")
				runtime.Goexit()
			}
		}
	}
}
//function to check if the server response idicates that there is 
//nothing else the server needs to tell the client
//returns true if there is nothing left the server needs to say, false otherwise
func checkResponse(response string) bool{
	for _,msg := range SERVER_MESSAGES{
		if response == msg{
			return true
		}
	}
	return false
}
//function that returns the backend server that the user will communicate with 
func selectServer() net.Conn{
	for i := 0; i < DEFAULT_SERVER_NUMBER; i++{
		port := strconv.Itoa(9001 + i)
		backend, berr := net.Dial("tcp", ":" + port)
		if berr == nil{
			fmt.Println("Using " + port)
			return backend
		}
	}
	return nil
}
