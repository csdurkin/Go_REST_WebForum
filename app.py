"""
APP.PY: Defines API (Three parts: endpoints, data shape, behaivour)
DIFFERENCES:    1. Endpoints - which URLS and functions exist, python: @app.route, go: mux.HandleFunc()
                2. Data shape - what json looks like, python: implicit in dict, go: structs with json tags
                3. Behaivour - what happens on each call, python - the function body, go: handle function body
"""
#TODO: gofmt: how do we import and use? 



"""
GO PORT: STRUCT MAP
Where each Go struct comes from in this file.

Stored data
  User               line 166   user_data = {...}
  Post               line 50    post_data = {...}
  Store              lines 8-10 posts, users, lock

Request bodies (look for data.get / data[...])
  CreateUserRequest  lines 146-147
  EditUserRequest    lines 207-209
  CreatePostRequest  lines 20-23

Responses
  Reply / PostResponse  lines 79-81  (replies list)
  Error                 every jsonify({'err': ...})

Not structs (URL query params, read with request.args)
  delete, search by datetime, search by user
"""




from flask import Flask, request, jsonify
from secrets import token_urlsafe
from datetime import datetime, timezone
import threading

app = Flask(__name__)


posts = {}                                                  #Dictonary to store posts and associated data
users = {}                                                  #Dictionary to store users and associated data
lock = threading.Lock()                                     #Creates a lock object; synchronizes access to resources

"""
GO IMPLEMENTATION

NOTES: 
    - Struct tags: backticks (not apostrophes!) map Go field names to JSON keys. Keep API's JSON identical to python version.
    - Stucture: {name type tag}. White space between these is flexible. Don't do this by hand though - use go fmt. 
"""

#*STORED DATA*

#RELATED PYTHON: store_post, post_data = {'id': post_id, 'key': key, 'timestamp': timestamp, 'msg': msg, 'user_id': user_id, 'post_parentid': post_parentid}
type Post struct {
    ID              int         `json:"id"`
    Key             string      `json:"key"`
    Timestamp       time.Time   `json:"timestamp"`
    Msg             string      `json:"msg"`
    UserID          int         `json:"user_id"`
    ParentID        *int        `json:"post_parentid"`
}


# RELATED PYTHON: user_data = {'id': user_id, 'key': user_key, 'user_screenname': user_screenname, 'user_realname': user_realname}    
type User struct {
    ID              int         `json:"id"`                                     #Important: Backticks, not appostrophes. Using these matches the API's JSON (and previous python implmentation) 
    Key             string      `json:"key"`
    ScreenName      string      `json:"user_screenname"`
    RealName        string      `json:"user_realname"`
}


#*REQUEST BODY*

#*RESPONSES*



# ENDPOINT 1: POST
# EXTENSIONS INCLUDED: (1: USERS) user_id and user_key to show authors (3: REPLIES) post_parentid allows for threaded replies
    
@app.route("/post", methods=['POST'])
def store_post():
    
    try:
        data = request.get_json()                           #Pull JSON data from HTTP request
        msg = data['msg']                                   #Access value at key 'msg' and store
        user_id = data.get('user_id')                       #Pull user_id from request data
        user_key = data.get('user_key')                     #Pull user_key from request data
        post_parentid = data.get('post_parentid')           #Specify what post is being replied to

        if not user_id or not user_key:
            return jsonify({'err': 'User ID and user key are required.'}), 400

        with lock:                                              #Ensrues one thread can exectue the section at a time

            #Check if user exists
            user = None

            for existing_user in users.values():
                if str(existing_user['id']) == str(user_id) and str(existing_user['key']) == str(user_key):
                    user = existing_user
                    break
            
            if not user:
                return jsonify({'err': 'User not found.'}), 404

            #if user['key'] != user_key:
                #return jsonify({'err': 'Invalid user credentials. Flag'}), 403

            #Create post data
            post_id = len(posts) + 1                           
            key = token_urlsafe(20)                             #Generate random key; token_urlsafe preferred method over randbelow or math.random
            timestamp = datetime.utcnow().replace(tzinfo=timezone.utc).isoformat() #Define by current time in UTC timezone; set to ISO format

            #Store post
            post_data = {'id': post_id, 'key': key, 'timestamp': timestamp, 'msg': msg, 'user_id': user_id, 'post_parentid': post_parentid}
            posts[post_id] = post_data

    except KeyError:
        return jsonify({'err': 'Bad Request'}), 400         #400 Bad Request: Client-side error with request. Request not completed.
    
    except Exception as e:
        app.logger.error(str(e))
        return jsonify({'err': 'An error occurred while processing your request.'}), 500  # 500 Internal Server Error

    #Return JSON Object
    return jsonify(post_data)

# ENDPOINT 2: GET
# EXTENSIONS INCLUDED: (3: REPLIES) Returns IDs for posts replying to the gotten post

@app.route("/post/<int:post_id>", methods=['GET'])
def get_post(post_id):
    
    with lock:

        post_data = posts.get(post_id)
        
        if not post_data: 
            return jsonify({'err': 'Post not found.'}), 404     #404 Not Found: Client-side Error. Resource does not exists on server.

        replies = [post for post in posts.values() if post.get('post_parentid') == post_id]     #Pull posts that acts as reply to input's post_id
        
        if replies:
            post_data['replies'] = [{'Post_ReplyID': reply['post_id'], 'timestamp': reply['timestamp']} for reply in replies]
        else: 
            post_data['replies'] = 'No Replies'

        return jsonify(post_data)


# ENDPOINT 3: DELETE
    
@app.route('/post/<int:post_id>/delete/<string:key>', methods=['DELETE'])
def detele_post(post_id, key):

    user_id = request.args.get('user_id')
    user_key = request.args.get('user_key')
    
    try:
        data = request.get_json()

        with lock: 

            post_data = posts.get(post_id)

            if not post_data: 
                return jsonify({'err': 'Post not found.'}), 404 
                
            # Check if user exists
            user = users.get(str(user_id))

            if not user or user['key'] != user_key or user['key'] != post_data['key']:    
                return jsonify({'err': 'Invalid user credentials.'}), 403        

            if post_data['key'] != key: 
                return jsonify({'err': 'Forbidden.'}), 403  

            del posts[post_id]

            return jsonify(post_data)
        
    except KeyError:
        return jsonify({'err': 'Bad Request'}), 400 

    
        
 #Check if user exists
        user = users.get(str(user_id))

        if not user or user['key'] != user_key or user['key'] != post_data['key']:      #Add additional check that user key and key in post           
            return jsonify({'err': 'Invalid user credentials.'}), 403        

        if post_data['key'] != key: 
            return jsonify({'err': 'Forbidden.'}), 403          #403 Forbidden: Client has insufficient permissions 

        del posts[post_id]

        return jsonify(post_data)
    

#EXTENSION 1: USER
#EXTENSION 2: USER PROFILES

@app.route("/user", methods=['POST'])
def create_user():

    try:
        data = request.get_json()
        user_id = len(users) + 1  # Simple incrementing integer ID
        user_key = token_urlsafe(20)
        user_screenname = data.get('user_screenname')
        user_realname = data.get('user_realname')

    except KeyError:
        return jsonify({'err': 'Bad Request'}), 400       
    

    with lock: 

        #Check if a unique username is provided

        if not user_screenname:
            return jsonify({'err': 'Unique screenname needed to create user.'}), 400

        for existing_user in users.values():
            if 'user_screenname' in existing_user and existing_user['user_screenname'] == user_screenname:
             return jsonify({'err': 'Provided screenname is not unique.'}), 400

        #store user

        user_data = {'id': user_id, 'key': user_key, 'user_screenname': user_screenname, 'user_realname': user_realname}
        users[user_id] = user_data

    # Return JSON Object with user key
        
    return jsonify(user_data)


#EXTENSION 2 (Cont): User Profiles, Get Meta Data

@app.route("/user/<string:user_identifier>", methods=['GET'])
def get_user_metadata(user_identifier):
  
    with lock: 
        
        user_data = None
        
        #Check if matches 'id' key
        if user_identifier in users:
            user_data = users[user_identifier]

        else:
            
            for existing_user in users.values():
                
                if str(existing_user.get('id')) == user_identifier or existing_user.get('user_screenname') == user_identifier:
                    user_data = existing_user
                    break

        if not user_data: 
            return jsonify({'err': 'User not found.'}), 404
        
        return jsonify(user_data)
    
#EXTENSION 2 (Cont): User Profiles, Edit Meta Data
    
@app.route("/user/<string:user_identifier>/edit", methods=['PUT'])
def edit_user_metadata(user_identifier):

    try:
        data = request.get_json()                          
        user_key = data.get('user_key')
        new_user_realname = data.get('user_realname')
        new_user_screenname = data.get('user_screenname')
    
        with lock:
            
            user_data = users.get(user_identifier)

            # Check if the input is a user ID
            if user_identifier.isdigit():
                user_identifier = int(user_identifier)
                user_data = users.get(user_identifier)

            if not user_data:
                return jsonify({'err': 'User not found.'}), 404

            if not user_key or user_data['key'] != user_key:
                return jsonify({'err': 'Forbidden. [Flag]'}), 403

            if new_user_realname:
                user_data['user_realname'] = new_user_realname

            if new_user_screenname:
                user_data['user_screenname'] = new_user_screenname
                
    except KeyError:
        return jsonify({'err': 'Bad Request'}), 400

    return jsonify(user_data)


#GET ALL USERS
@app.route("/user", methods=['GET'])
def get_all_users():
    with lock:
        return jsonify(list(users.values()))


# EXTENSION 4: DATE/TIME SEARCH

@app.route("/posts/search/datetime", methods=['GET'])
def search_datetime():
    
    try:
        
        start_datetime_str = request.args.get('start_datetime')
        end_datetime_str = request.args.get('end_datetime')
        posts_inrange = []
        
        if start_datetime_str:
            start_datetime = datetime.fromisoformat(start_datetime_str).replace(tzinfo=timezone.utc)  #Fixes issue of comparing datetimes with different offset awareness
        else:
            start_datetime = None
            
        if end_datetime_str:
            end_datetime = datetime.fromisoformat(end_datetime_str).replace(tzinfo=timezone.utc)
        else:
            end_datetime = None
        
        for post_data in posts.values():
            
            post_timestamp = (post_data['timestamp'])
            
            if isinstance(post_timestamp, str):
                post_timestamp = datetime.fromisoformat(post_timestamp)

            # Check if post's timestamp is within range of user's timeframe (sets to True)
            # or if user's time is not specified (sets to True)
            start_datetime_check = start_datetime is None or (isinstance(post_timestamp, datetime) and start_datetime <= post_timestamp)
            end_datetime_check = end_datetime is None or (isinstance(post_timestamp, datetime) and end_datetime >= post_timestamp)

            # Add posts to the array to be returned
            if start_datetime_check and end_datetime_check:
                posts_inrange.append(post_data)
        
        return jsonify(posts_inrange)

    except ValueError:
        return jsonify({'err': 'Invalid date/time format.'}), 400
        

#EXTENSION 5: USER SEARCH

@app.route("/posts/search/user", methods=['GET'])
def search_user(): 

    user_id_str = request.args.get('user_id')

    try:
        user_id = int(user_id_str)

    except (ValueError, TypeError):
        return jsonify({'err': 'Invalid user_id format.'}), 400

    with lock: 

        if user_id not in users:
            return jsonify({'err': 'User not found.'}), 404
        
        posts_byuser = []

        for post_data in posts.values():
            
            if post_data.get('user_id') == user_id:
                posts_byuser.append(post_data)

        return jsonify(posts_byuser)

            
# MAIN

if __name__ == '__main__':
    app.run(debug='True')
