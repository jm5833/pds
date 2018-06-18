Jack Martinez
N18522789

Twitter clone that has the ability to make accounts, login, logout, and delete 
user accounts. Users who are logged in can make posts, but not delete posts.
Users can also choose who they follow, and they can also choose to unfollow 
someone they no longer wish to see updates about

Run Notes:
	backend - go run backend.go
	frontend - go run frontend.go
	requires the backend to be started before the frontend
	the front end is on localhost:8000

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

