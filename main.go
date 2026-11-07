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
		"strconv"			// convert text to numbers, e.g. "3" → 3; Python Equivalent: int()
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
		writeJSON(writer, statusCode, ErrorResponse{Err: message})						//ErrorResponse: Struct defined above, defines shape of every error API sends
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

func storePost(writer http.ResponseWriter, request *http.Request) {

	// 1 Read JSON input
	var postRequest CreatePostRequest 									// order: variable declaration, var. name, type (here, struct type CreatePostRequest)
	err := json.NewDecoder(request.Body).Decode(&postRequest)			// Reverse of encoder, notes in helper:writeJSON. Read the JSON bdy and store at the pointer of requestBody

	// 2 Error Handling
	
	if err != nil || postRequest.Msg == "" || postRequest.UserID == 0 || postRequest.UserKey == "" {
			writeError(writer, http.StatusBadRequest, "Bad Request")
			return
		}

	// 3 Generate Secret Key
	postKey, err := newKey()
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "Error processing request (key generation).")
		return
	}

	// 4 Lock before touching shared data 
	store.mu.Lock()													// Go (sync.Mutex): blocks other requests from the shared data until unlocked; Python: with lock:
	defer store.mu.Unlock()											// Go keyword: runs Unlock when the function exits, however it exits; Python: the end of the with block

	// 5 Check user exists, key matches 
	user, found := store.users[postRequest.UserID]
	if !found || user.Key != postRequest.UserKey {
		writeError(writer, http.StatusNotFound, "User not found.")
		return
	}

	// 6 Assign next ID to post and store post 
	store.nextPostID++												// Set in store struct
	post := &Post{
		ID:			store.nextPostID, 
		Key: 		postKey,
		Timestamp: 	time.Now().UTC(),
		Msg:		postRequest.Msg, 
		UserID:		postRequest.UserID, 
		ParentID: 	postRequest.ParentID,
	}
	store.posts[post.ID] = post										//Reminder: key is post's ID, value is its pointer

	writeJSON(writer, http.StatusOK, post)

}


// GET /post/{id}
// Python: get_post
func getPost(writer http.ResponseWriter, request *http.Request) {

	// 1 Read input: post ID from URL path
	postID, err := strconv.Atoi(request.PathValue("id"))			// strconv - string converation; Atoi - ASCII to integer
	
	// 2 Error handling: reject a post ID that isn't a number (400)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "Bad Request")
		return
	}

	// 3 Lock before touching shared data (Python: with lock:)
	store.mu.Lock()
	defer store.mu.Unlock()

}

// DELETE /post/{id}/delete/{key}
// Python: delete_post
func deletePost(writer http.ResponseWriter, request *http.Request) {}

// POST /user
// Python: create_user
func createUser(writer http.ResponseWriter, request *http.Request) {
	
	// 1 Read input: decode JSON body into userRequest
	var userRequest CreateUserRequest
	err := json.NewDecoder(request.Body).Decode(&userRequest)

	// 2 Error handling: reject bad JSON or missing screen name (400)
	if err != nil || userRequest.ScreenName == "" {
		writeError(writer, http.StatusBadRequest, "Bad Request")
		return
	}

	// 3 Generate secret key
	userKey, err := newKey()
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "Error processing request (key generation).")
		return
	}

	// 4 Lock before touching shared data (Python: with lock:)
	store.mu.Lock()
	defer store.mu.Unlock()

}

// GET /user/{identifier}
// Python: get_user_metadata
func getUserMetadata(writer http.ResponseWriter, request *http.Request) {
	
	// 1 Read input: user ID or screen name from URL path
	identifier := request.PathValue("identifier")

	// 2 Error handling: none needed (any text is a valid identifier; not found is checked in step 5)

	// 3 Lock before touching shared data (Python: with lock:)
	store.mu.Lock()
	defer store.mu.Unlock()

}

// PUT /user/{identifier}/edit
// Python: edit_user_metadata
func editUserMetadata(writer http.ResponseWriter, request *http.Request) {}

// GET /users
// Python: get_all_users
func getAllUsers(writer http.ResponseWriter, request *http.Request) {

	// 1 Read input: not needed
	// 2 Error handling: none needed

	// 3 Lock before touching shared data (Python: with lock:)
	store.mu.Lock()
	defer store.mu.Unlock()

}

// GET /posts/search/datetime
// Python: search_datetime
func searchDatetime(writer http.ResponseWriter, request *http.Request) {

	// 1 Read input: start and end times from query string. Notes on why query and not pathvalue in search_user
	startText := request.URL.Query().Get("start_datetime")
	endText := request.URL.Query().Get("end_datetime")

	// 2 Error handling: parse times; reject any provided time that isn't ISO 8601 (400)
	// time.Parse (text → time.Time) using RFC3339 (ISO 8601 layout, e.g. 2023-01-01T00:00:00Z); returns (time.Time, error), error if text doesn't match
	startTime, startErr := time.Parse(time.RFC3339, startText)
	endTime, endErr := time.Parse(time.RFC3339, endText)
	if (startText != "" && startErr != nil) || (endText != "" && endErr != nil) {
		writeError(writer, http.StatusBadRequest, "Invalid date/time format.")
		return
	}

	// 3 Lock before touching shared data (Python: with lock:)
	store.mu.Lock()
	defer store.mu.Unlock()

}

// GET /posts/search/user (expample: /posts/search/user?user_id=1)
// Python: search_user
func searchUserPosts(writer http.ResponseWriter, request *http.Request) {

	// 1 UserId
	userId, err := strconv.Atoi(request.URL.Query().Get("user_id"))	//Why not pathValue? The user_id is placed in a querystring, while the post_id was directly in the url path
		
	// 2 Error handling: reject a missing or non-numeric user ID (400)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "Invalid user_id format.")
		return
	}

	// 3 Lock before touching shared data (Python: with lock:)
	store.mu.Lock()
	defer store.mu.Unlock()

}

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
	mux.HandleFunc("GET /users", getAllUsers)
	mux.HandleFunc("GET /posts/search/datetime", searchDatetime)
	mux.HandleFunc("GET /posts/search/user", searchUserPosts)

	// 3. Start the server (Python: app.run())
	log.Println("Listening on http://127.0.0.1:5000")
	log.Fatal(http.ListenAndServe("127.0.0.1:5000", mux))
}


/* GO NOTES: 
- Make sure to use ":=" not just ":" or ever "==": this intiates and sets the variable
- Watch CAPS: capitalization matters in GO 
- Pointers are more declared in GO than Python. Learn more about this part of syntax.
- strconv.Atoi: version of int(); more written out conversions in GO 
*/