Jack Martinez
N18522789

Twitter clone that has the ability to make accounts, login, logout, and delete 
user accounts. Users who are logged in can make posts, but not delete posts.
Users can also choose who they follow, and they can also choose to unfollow 
someone they no longer wish to see updates about

Run Notes:
	backend - go run backend.go
	frontend - go run frontend.go
	the backend needs to be started before the frontend
	the frontend runs on localhost:8000
	

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
	BadUP								-bad username/password combination
	LoginOK								-sucessful login
	PostOK								-sucessful post
	End									-end of posts for the user

File Data
	user.txt - holds the username/password combinations
	post.txt - holds all posts made by all users
	follow.txt - holds the list of people the user follows

Data format
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

