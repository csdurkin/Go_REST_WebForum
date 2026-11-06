/*
GO PORT of app.py

FILE ORDER
1. Imports
2. Structs: stored data, request bodies, responses
3. Store Instance: in-memory data and its lock
4. Helpers 
5. Handlers, one per endpoint
6. main(): routes and server start
*/

package main

// 1 IMPORTS -----------------------------------------------------------------------------

	import (	
		"log"		// startup and error messages
		"net/http"	// server, routes, ResponseWriter, REquest
		"sync"		// Python Equivalent: threading
		"time"		// Python Equivalent: datetime, timezone
	)

// 2 STRUCTS -----------------------------------------------------------------------------

	// 2A STORED DATA

		//Store: in-memory database
		//Python Equivalent: posts = {}, users = {}, lock = threading.Lock()
		type Store struct {
			mu    sync.Mutex				// lock: hold whenever reading or changing; Mutex (mutual exclusion) lock that lets only one access to shared data at a time
			users map[int]*User				// * - Map stores pointers, not copies of them
			posts map[int]*Post
			nextUserID int
			nextPostID int
		}

		type Post struct {
			ID        int       `json:"id"`
			Key       string    `json:"key"`
			Timestamp time.Time `json:"timestamp"`
			Msg       string    `json:"msg"`
			UserID    int       `json:"user_id"`
			ParentID  *int      `json:"post_parentid"`
		}

		type User struct {
			ID         int    `json:"id"`
			Key        string `json:"key"`
			ScreenName string `json:"user_screenname"`
			RealName   string `json:"user_realname"`
		}

	// 2B REQUEST DATA (JSON FROM CLIENTS)

		type CreatePostRequest struct {
			Msg      string `json:"msg"`
			UserID   int    `json:"user_id"`
			UserKey  string `json:"user_key"`
			ParentID *int   `json:"post_parentid"`
		}

		type CreateUserRequest struct {
			ScreenName string `json:"user_screenname"`
			RealName   string `json:"user_realname"`
		}

		type EditUserRequest struct {
			UserKey    string `json:"user_key"`
			RealName   string `json:"user_realname"`
			ScreenName string `json:"user_screenname"`
		}

	//2C RESPONSES (JSON SERVERS SENDS, BEYOND USER OR POST)

		type Reply struct {
			PostReplyID int       `json:"Post_ReplyID"`
			Timestamp   time.Time `json:"timestamp"`
		}

		type PostResponse struct {
			Post
			Replies []Reply `json:"replies"`			//[]Reply: A list of replies
		}

		type ErrorResponse struct {
			Err string `json:"err"`
		}

// 3 STORE INSTANCE -----------------------------------------------------------------------------

	//Actual instance of store to hold all forum data while server runs
	//Notes: maps must be created before use, writing to nil map crashes app; mu and counters need no setup, zero values are ready

var store = &Store{							// &: ampersand means "the address of"; this creates a store and gives a pointer to it; ensures the original is shared with a function
	users: map[int]*User{},					// {}: creates an empty map. Without, only a type, not an actual maps
	posts: map[int]*Post{},					// How to read line: new, empty map that will store key (int) and value (pointer to a Post); Note: all lines need trailing commas in multi-line struct
}

// 4 HELPERS -----------------------------------------------------------------------------

	// TODO: CREATE A HELPER VERSION OF JSONIFY, AS NOT AVAILABLE IN GO

// 5 HANDLERS (ONE PER ENDPOINT) ------------------------------------------------

/*
NOTES: HANDLER SYNTAX

	- func: declares function, python equivalent: def
	- functionName: use camelCase
	- w http.ResponderWriter: 
		- ResponderWriter: built into Go's net/http packacke. Connaction back to client. Whatever writen to w is sent as response. 
		- w: where the response is written. Name first, type second.
		- This replaces Python's jsonify(...)
	- r *http.Request: 
		- http.Request: built into Go's net/http package. Holds everythign fclient sent, passed to handler as r.
		- r: the cincoming request (body, url, query params)
		- *: pointer to the request body, not a copy
		- This replaces Python's request object


*/

// POST /post
// Python: store_post
func storePost(w http.ResponseWriter, r *http.Request) {}

// GET /post/{id}
// Python: get_post
func getPost(w http.ResponseWriter, r *http.Request) {}

// DELETE /post/{id}/delete/{key}
// Python: detele_post
func deletePost(w http.ResponseWriter, r *http.Request) {}

// POST /user
// Python: create_user
func createUser(w http.ResponseWriter, r *http.Request) {}

// GET /user/{identifier}
// Python: get_user_metadata
func getUserMetadata(w http.ResponseWriter, r *http.Request) {}

// PUT /user/{identifier}/edit
// Python: edit_user_metadata
func editUserMetadata(w http.ResponseWriter, r *http.Request) {}

// GET /user
// Python: get_all_users
func getAllUsers(w http.ResponseWriter, r *http.Request) {}

// GET /posts/search/datetime
// Python: search_datetime
func searchDatetime(w http.ResponseWriter, r *http.Request) {}

// GET /posts/search/user
// Python: search_user
func searchUser(w http.ResponseWriter, r *http.Request) {}

// 6 MAIN

// 6 MAIN: ROUTES AND SERVER START ----------------------------------------------

func main() {
	
	//Create the router (Python: app = Flask(__name__))
	mux := http.NewServeMux()

	//Register routes (Python: the @app.route lines)
	mux.HandleFunc("POST /post", storePost)
	mux.HandleFunc("GET /post/{id}", getPost)
	mux.HandleFunc("DELETE /post/{id}/delete/{key}", deletePost)
	mux.HandleFunc("POST /user", createUser)
	mux.HandleFunc("GET /user/{identifier}", getUserMetadata)
	mux.HandleFunc("PUT /user/{identifier}/edit", editUserMetadata)
	mux.HandleFunc("GET /user", getAllUsers)
	mux.HandleFunc("GET /posts/search/datetime", searchDatetime)
	mux.HandleFunc("GET /posts/search/user", searchUser)

	// 3. Start the server (Python: app.run())
	log.Println("Listening on http://127.0.0.1:5000")
	log.Fatal(http.ListenAndServe("127.0.0.1:5000", mux))
}