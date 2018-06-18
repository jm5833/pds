Jack Martinez
N18522789

Twitter clone that has the ability to make accounts, login, logout, and delete 
user accounts. Users who are logged in can make posts, but not delete posts.
Users can also choose who they follow, and they can also choose to unfollow 
someone they no longer wish to see updates about

Run Notes:
	create 3 folders(I did Server0, Server1, and Server2 for testing)
	run in the following order:

	replication server - go run replication.go
	server select - go run selector.go
	backend - go run backend.go [folder directory] [port]
		[folder directory] - the directory you want this particular 
							 backend to store it's files on
		[port]			   - the port you want this backend to run on
						     starting from 9001
	frontend - go run frontend.go
	the backend needs to be started before the frontend
	the frontend runs on localhost:8000
	
	The server selector will only handle up to 3 servers because of the DEFAULT_SERVER_NUMBER
	While more backend servers can connect to it, the server select will never select them as
	the server to connect the client to
	

types of messages the backend is expecting
	Register;[user];[passowrd]			-creates a user with the username/password provided
	Login;[user];[password]				-attempts to log a user in with the username/password given
	Post;[user];[content];[timestamp]	-creates a post
	GetPost;[user]						-gets all posts from people the user follows
	Follow;[user];[a];[b]...			-adds users a,b,... to user's follow list
	Unfollow;[user];[a];[b]...			-removes a,b,... from the user's follow list
	Delete;[user]						-deletes the user's profile
	GetFollowed;[user]					-gets all people the user follows
	GetNotFollowed;[user]				-gets all the people the user doesn't follow

types of messages the client expects
	NewUser 							-new successful user
	UserExists							-couldn't create the username because it already exists
	BadAL								-bad argument list
	LoginOK								-sucessful login
	PostOK								-sucessful post
	End									-marks the end of the list of items being sent 

File Data
	user.txt - holds the username/password combinations
	post.txt - holds all posts made by all users
	follow.txt - holds the list of people the user follows

Data format within each file
	user.txt - [username],[password]
	post.txt - [username],[content],[timestamp]
	follow.txt - [username],[person followed],[person followed],...

Shared Data
	user.txt, post.txt, and follow.txt are the 3 files that each goroutine
	accesses. I address potential race conditions by having a specific
	lock for each file. When a go routine needs to access 1 or more files,
	it obtains the locks it needs in a specific order, and releases each 
	lock in a specific order as well. 

Lock ordering
	Locks are always obtained in the following order:
	user_lock -> post_lock -> follow_lock
	
	Locks are always released in the following order:
	follow_lock -> post_lock -> user_lock

	Eg. if a function needs the user file and the follow file, it will lock
	the user_lock lock first, then the follow_lock lock second, and then 
	release the follow_lock first, then user_lock second
Replication
	Replication is handled by the replication.go file
	Whenever a backend server does an operation that writes to a file, it sends that 
	operation to the replication server. From there, the replication server sends out
	a message to all other connected servers telling them to update the changed files.
	
	e.g Server0 has a user register with Register;user;user
		Server0 sends Register;user;user to the replication server
		The replication server then sends Register;user;user to all other servers
Server Failure
	When a server fails(or is force quit), selector.go is supposed to reroute you to
	another available server. 
Notes:
	selector.go doesn't work properly, when the server you're connected on dies, it does
	manage to find another working server for you, and data will be transfered from client to
	server, but data from server to client doesn't work. However, if a server dies, and you 
	stop/start frontend.go again, the selector will get you on a working server and data 
	transfer will work properly.
	
	Sometimes when accessing /home and /profile, it'll display no/incorrect information. This
	tends to happen only on redirects, and reloading the page shows the correct information
