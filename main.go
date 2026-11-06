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
		"log"				// startup and error messages
		"net/http"			// server, routes, ResponseWriter, REquest
		"sync"				// Python Equivalent: threading
		"time"				// Python Equivalent: datetime, timezone
		"crypto/rand"     	// secure random bytes for keys; Python Equivalent: secrets
		"encoding/base64" 	// turns key bytes into URL-safe text; Python Equivalent: secrets.token_urlsafe
		"encoding/json"   	// read and write JSON; Python Equivalent: request.get_json, jsonify
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

	//FUNCTION: writeJSON: sends data as JSON with status code; Python equivalent: return jsonify(data), statusCode
	func writeJSON(writer http.ResponseWriter, statusCode int, data any) {				// any: built-in Go type meaning "a value of any type"
		writer.Header().Set("Content-Type", "application/json")						// Returns response's headers, adds one header telling the client the body will be JSON
		writer.WriteHeader(statusCode)													// method on http.ResponseWriter to set status code
		json.NewEncoder(writer).Encode(data)											// json.NewEncoder(writer): creates JSON encord, an object that writers JSON to writer (sets up where output goes); Encoder: converts data into Json text and writers to destination 
	}	
	
	//FUNCTION: writeError: sends {"err": message} with given status code; Python equivalent: return jsonify({'err}: message), statusCode
	func writeError(writer http.ResponseWriter, statusCode int, message string) {
		writeJSON(writer, statusCode, ErrorResponse{Err: message})						//rrorResponse: Struct defined above, defines shape of every error API sends
	}

	//FUNCTION: newKey: returns a random URL-safe key; python equivalent: token_urlsafe(20)
	func newKey() (string, error){														//Why two ()(): In GO, function signatures have two sets of brackets, one for inputs, the second for outputs. 
		randomBytes := make([]byte, 20)													// Make list of 20 bytes set to 0; :=  - Creates a new variable and gives it value in one step
		_, err := rand.Read(randomBytes)												// _: must be used, GO refuses unused variables; Fill with random and secure bytes; rand (crypto/rand pack); Read: package function that produces random bytes
		if err != nil {									
			return "", err																// Why "", err: newKey promises to return two values, a string and an error
		}
		key := base64.RawURLEncoding.EncodeToString(randomBytes)						// base64 (package, way of writing raw texs as 64 safe characters, A-Z, a-z, 0-9) . RawURLEncoding (URL-safe, no padding) . EncodeToString (bytes → text)
		return key, nil
	}


// 5 HANDLERS (ONE PER ENDPOINT) ------------------------------------------------

/*
NOTES: HANDLER SYNTAX

	- func: declares function, python equivalent: def
	- functionName: use camelCase
	- writer http.ResponderWriter: 
		- ResponderWriter: built into Go's net/http packacke. Connaction back to client. Whatever writen to writer is sent as response. 
		- w: where the response is written. Name first, type second.
		- This replaces Python's jsonify(...)
	- request *http.Request: 
		- http.Request: built into Go's net/http package. Holds everythign fclient sent, passed to handler as r.
		- request: the cincoming request (body, url, query params)
		- *: pointer to the request body, not a copy
		- This replaces Python's request object


*/

// POST /post
// Python: store_post
func storePost(writer http.ResponseWriter, request *http.Request) {}

// GET /post/{id}
// Python: get_post
func getPost(writer http.ResponseWriter, request *http.Request) {}

// DELETE /post/{id}/delete/{key}
// Python: detele_post
func deletePost(writer http.ResponseWriter, request *http.Request) {}

// POST /user
// Python: create_user
func createUser(writer http.ResponseWriter, request *http.Request) {}

// GET /user/{identifier}
// Python: get_user_metadata
func getUserMetadata(writer http.ResponseWriter, request *http.Request) {}

// PUT /user/{identifier}/edit
// Python: edit_user_metadata
func editUserMetadata(writer http.ResponseWriter, request *http.Request) {}

// GET /user
// Python: get_all_users
func getAllUsers(writer http.ResponseWriter, request *http.Request) {}

// GET /posts/search/datetime
// Python: search_datetime
func searchDatetime(writer http.ResponseWriter, request *http.Request) {}

// GET /posts/search/user
// Python: search_user
func searchUser(writer http.ResponseWriter, request *http.Request) {}

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