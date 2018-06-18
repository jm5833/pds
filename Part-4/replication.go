/*
	Jack Martinez
	Semester Project - Part 4
	N18522780
	Replication Server
*/
package main

import(
	"fmt"
	"log"
	"net"
	"bufio"
	"sync"
	"strings"
	"runtime"
)
//list of connections
var connectionList []net.Conn
//list of all messages
var allMessages []string
//lock for the connectionList and allMessages list
var connectionsLock = &sync.Mutex{}
var messagesLock = &sync.Mutex{}

func main(){
	server, serr := net.Listen("tcp", ":8999")
	checkError(serr, 1)
	acceptBackend(server)
}

//function to check error return values
func checkError(err error, checkType int){
	switch checkType{
	case 0:
		if err != nil{
			log.Println(err)
		}
	case 1:
		if err != nil{
			panic(err)
		}
	}
}

//function to accept the backend server connections
func acceptBackend(server net.Listener){
	fmt.Println("Ready to accept backend connections.")
	for{
		backend, err := server.Accept()
		checkError(err, 1)
		connectionsLock.Lock()
		connectionList = append(connectionList, backend)
		connectionsLock.Unlock()
		go handleBackend(backend)
	}
}
//fucntion to handle the backend
func handleBackend(backend net.Conn){
	updateServer(backend)
	backbuf := bufio.NewScanner(backend)
	counter := 0
	flag := false
	var message string
	for{
		message = ""
		backbuf.Scan()
		fmt.Println("server input recieved")
		message = backbuf.Text()
		if len(message) == 0{
			flag = true
			counter++
		}
		if flag && counter >= 3{
			runtime.Goexit()
		}else if flag{
			continue
		}
		fmt.Println("|" + message +"|")
		sendToAll(message, backend)
	}
}
//function to update the server for when it connects to the replication
//server for the first time
func updateServer(backend net.Conn){
	messagesLock.Lock()
	fmt.Println("Initial update")
	defer messagesLock.Unlock()
	for i := 0; i < len(allMessages); i++{
		message := allMessages[i]
		fmt.Println(message)
		backend.Write([]byte(message + "\n"))
	}
	fmt.Println("Finished initial update")
}
//function to send all backends the info on how to update their files
func sendToAll(message string, backend net.Conn){
	connectionsLock.Lock()
	messagesLock.Lock()
	defer messagesLock.Unlock()
	defer connectionsLock.Unlock()

	index := strings.Index(message, ";")
	if index == -1{
		return
	}
	servernum := getServerNum(backend)
	for i := 0; i < len(connectionList); i++{
		//we skip the server that called this function to avoid sending them 
		//a redundant update message
		if i == servernum{
			continue
		}
		retval, err := connectionList[i].Write([]byte(message + "\n"))
		if err != nil && retval == 0{
			connectionList = append(connectionList[:i],connectionList[i+1:]...)
		}
	}
	allMessages = append(allMessages, message)
}
//function to return the index of backend in connectionList
func getServerNum(backend net.Conn) int{
	index := -1
	for i,server := range connectionList{
		if server == backend{
			index = i
			return index
		}
	}
	return index
}
