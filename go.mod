type Store struct {             
    my sync.Mutex               
    users map[int]*User         
    posts map[int]*Post
}

type Post struct {
    ID int `json:"id`"
    Key string `json:"key"`
    Timestamp time.Time `json:"timestamp"
    Msg string `json:"msg"`
    UserID int `json:"user_id"`
    ParentID *int `json:"post_parentid"
}

type User struct {
    ID int `json: "id"`
    Key string `json:"key"`
    ScreenName  string  `json:"user_screenname"
    RealName    string  `json:"user_realname"
}

type CreatePostRequest struct {
    Msg string "json:msg"
    User_id int "json:user_id"
    User_key int "json:user_key"
    post_parentid *int "json:post_parentid"             
}

type CreateUserRequest struct {
    User_screenname string "json:user_screenname"
    User_realname string "json:user_realname"
}

type EditUserRequest struct {
    User_key int "json:user_key"
    New_user_realname string "json:user_realname"
    New_user_screenname "json:user_screenname"
}

type Reply struct {
	PostReplyID int    `json:"Post_ReplyID"`
	Timestamp   string `json:"timestamp"`
}

type PostResponse struct {
	Post            
	Replies []Reply `json:"replies"`
}