type Store struct {
	mu    sync.Mutex
	users map[int]*User
	posts map[int]*Post
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

type Reply struct {
	PostReplyID int       `json:"Post_ReplyID"`
	Timestamp   time.Time `json:"timestamp"`
}

type PostResponse struct {
	Post
	Replies []Reply `json:"replies"`
}

type ErrorResponse struct {
	Err string `json:"err"`
}