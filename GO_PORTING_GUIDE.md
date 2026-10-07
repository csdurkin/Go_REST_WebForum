# Go Porting Guide

> **AI-generated document.** This guide was written by **Claude (Opus 5.5, medium effort)**, an AI model by Anthropic, at Connor Durkin's request in October 2026. It is a learning aid for this port, not original work by the repository owner. Code snippets were compiled and tested against Go 1.24.

The main concepts to understand when moving this forum API from Python (Flask) to Go. Each section explains how Go handles something, how it differs from the Python version, and where it shows up in this project.

This guide teaches concepts. It is not a step-by-step build.

---

## Contents

1. [Project setup](#1-project-setup)
2. [Static types and structs](#2-static-types-and-structs)
3. [Maps and missing values](#3-maps-and-missing-values)
4. [Values, pointers and mutation](#4-values-pointers-and-mutation)
5. [Errors instead of exceptions](#5-errors-instead-of-exceptions)
6. [HTTP routing with the standard library](#6-http-routing-with-the-standard-library)
7. [Reading and writing JSON](#7-reading-and-writing-json)
8. [Concurrency and locking](#8-concurrency-and-locking)
9. [Time and dates](#9-time-and-dates)
10. [Generating secure keys](#10-generating-secure-keys)
11. [Testing](#11-testing)
12. [Python to Go quick reference](#12-python-to-go-quick-reference)

---

## 1. Project setup

Go projects are organised as **modules**. A module is a folder with a `go.mod` file that names it and pins the Go version.

```bash
go mod init github.com/csdurkin/Go_REST_WebForum
go run .        # compile and run
go build        # produce a binary
go test ./...   # run all tests
go fmt ./...    # format code (Go has one official style)
```

| Python | Go |
|---|---|
| `pip install flask` | Not needed. The standard library covers HTTP and JSON |
| `requirements.txt` | `go.mod` and `go.sum` |
| Interpreted at runtime | Compiled to a single binary |
| `if __name__ == '__main__':` | `func main()` in `package main` |

**Version note:** use Go 1.22 or later. That release added method and path-parameter routing to the standard library, which removes the need for a framework.

---

## 2. Static types and structs

Python dictionaries accept any key and any value. Go requires you to declare the shape of your data up front with a **struct**.

```go
type User struct {
    ID         int    `json:"id"`
    Key        string `json:"key"`
    ScreenName string `json:"user_screenname"`
    RealName   string `json:"user_realname"`
}

type Post struct {
    ID        int       `json:"id"`
    Key       string    `json:"key"`
    Timestamp time.Time `json:"timestamp"`
    Msg       string    `json:"msg"`
    UserID    int       `json:"user_id"`
    ParentID  *int      `json:"post_parentid"`
}
```

Key ideas:

- **Struct tags** (the text in backticks) map Go field names to JSON keys. They keep the API's JSON identical to the Python version.
- **Capitalised fields are exported.** The JSON encoder can only see fields that start with a capital letter. A lowercase field is silently skipped.
- **Nullable values need pointers.** `post_parentid` can be `null`. A plain `int` cannot be null; its zero value is `0`. A `*int` can be `nil`, which encodes as `null`.
- **Types are checked at compile time.** In the Python version, `user_id` could arrive as a string or an integer, which caused comparison bugs. In Go, decoding fails if the type is wrong.

**Zero values:** every Go type has a default. `int` is `0`, `string` is `""`, `bool` is `false`, and pointers, maps and slices are `nil`. There is no `None` for basic types.

---

## 3. Maps and missing values

Python's `posts = {}` becomes a typed map.

```go
posts := map[int]Post{}
users := map[int]User{}
```

Reading a missing key does not raise an error. It returns the zero value. Use the **comma-ok idiom** to tell "missing" from "empty":

```go
post, ok := posts[id]
if !ok {
    // not found: return 404
}
```

| Python | Go |
|---|---|
| `posts.get(post_id)` returns `None` | `posts[id]` returns a zero-value `Post` |
| `if not post_data:` | `if !ok {` |
| `del posts[post_id]` | `delete(posts, id)` |
| `len(posts)` | `len(posts)` |
| `for p in posts.values():` | `for _, p := range posts {` |

**Map iteration order is random in Go.** Python dicts keep insertion order. If an endpoint returns a list, such as search results, sort it if order matters.

**ID generation:** the Python version uses `len(posts) + 1`, which reuses IDs after a delete. A separate counter that only increases avoids this.

---

## 4. Values, pointers and mutation

This is the biggest mental shift from Python.

In Python, everything is a reference. Changing a dict you got from `posts.get()` changes the stored post. In Go, **structs are copied** when assigned or returned from a map.

```go
p := posts[1]
p.Msg = "edited"   // changes the copy only
posts[1] = p       // write it back to store the change
```

Two options:

| Approach | Map type | Behaviour |
|---|---|---|
| Store values | `map[int]Post` | Safe by default; you must write changes back |
| Store pointers | `map[int]*Post` | Changes affect the stored item, as in Python |

**Why this matters here:** the Python `GET /post/<id>` endpoint adds a `replies` field to the stored post by accident, because it edits the shared dict. Go's copy semantics make that mistake harder to make.

Rule of thumb:

- `&x` gives the address of `x` (a pointer)
- `*p` gives the value a pointer points to
- Methods with pointer receivers (`func (s *Store) Add(...)`) can change the receiver; value receivers cannot

---

## 5. Errors instead of exceptions

Go has no `try/except`. Functions that can fail return an `error` as their last value, and you check it immediately.

```go
id, err := strconv.Atoi(r.PathValue("id"))
if err != nil {
    writeError(w, http.StatusBadRequest, "Invalid post ID.")
    return
}
```

| Python | Go |
|---|---|
| `try: ... except KeyError:` | `value, ok := m[key]` |
| `try: int(x) except ValueError:` | `n, err := strconv.Atoi(x)` |
| `raise` | `return ..., err` or `errors.New("...")` |
| Exceptions bubble up | Errors are returned and checked explicitly |
| `except Exception as e:` (catch-all) | No equivalent; handle each error where it occurs |

The pattern is **check, handle, return early**. It is verbose, but every failure point is visible. A small helper that writes JSON errors keeps handlers short.

`panic` exists but is for unrecoverable bugs, not normal control flow. The HTTP server recovers from panics in handlers, but you should not rely on that.

---

## 6. HTTP routing with the standard library

Flask decorators become handler functions registered on a **ServeMux**.

```go
mux := http.NewServeMux()
mux.HandleFunc("POST /post", createPost)
mux.HandleFunc("GET /post/{id}", getPost)
mux.HandleFunc("DELETE /post/{id}/delete/{key}", deletePost)

log.Fatal(http.ListenAndServe("127.0.0.1:5000", mux))
```

Every handler has the same signature:

```go
func getPost(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")            // path parameter
    userID := r.URL.Query().Get("user_id") // query string
    // ...
}
```

| Flask | Go |
|---|---|
| `@app.route("/post", methods=['POST'])` | `mux.HandleFunc("POST /post", fn)` |
| `<int:post_id>` | `{id}` then `strconv.Atoi(r.PathValue("id"))` |
| `request.args.get('x')` | `r.URL.Query().Get("x")` |
| `request.get_json()` | `json.NewDecoder(r.Body).Decode(&v)` |
| `return jsonify(data), 404` | set header, `w.WriteHeader(404)`, encode body |
| Wrong method returns 405 automatically | Same, when the pattern includes a method |

Path values are always strings. Converting them to integers is your job, which is where Flask's `<int:...>` did the work for you.

**Order matters when writing a response:** set headers first, then the status code, then the body. Once the body starts, the status is fixed at 200.

---

## 7. Reading and writing JSON

The `encoding/json` package converts between JSON and structs using the struct tags from section 2.

**Decoding a request:**

```go
var req struct {
    Msg      string `json:"msg"`
    UserID   int    `json:"user_id"`
    UserKey  string `json:"user_key"`
    ParentID *int   `json:"post_parentid"`
}
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    writeError(w, http.StatusBadRequest, "Bad Request")
    return
}
```

**Encoding a response:**

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}
```

Points to know:

- **Missing fields are not errors.** A missing `msg` decodes as `""`. You must check required fields yourself.
- **Unknown fields are ignored** unless you call `decoder.DisallowUnknownFields()`.
- **Anonymous structs** (declared inline, as above) are handy for request bodies that differ from stored data.
- **Content type is not enforced.** Flask's `get_json()` returned HTTP 415 when no JSON body was sent, which broke the Python delete endpoint. Go only reads the body if you ask it to.
- **Hiding fields:** a tag of `json:"-"` excludes a field from output. That is useful for secret keys.

---

## 8. Concurrency and locking

Go's `net/http` server runs **each request in its own goroutine** (a lightweight thread). Several handlers can touch the maps at once, so shared state needs a lock, as in the Python version.

```go
type Store struct {
    mu     sync.Mutex
    posts  map[int]Post
    users  map[int]User
    nextID int
}

func (s *Store) GetPost(id int) (Post, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    p, ok := s.posts[id]
    return p, ok
}
```

| Python | Go |
|---|---|
| `lock = threading.Lock()` | `var mu sync.Mutex` |
| `with lock:` | `mu.Lock()` then `defer mu.Unlock()` |
| Threads | Goroutines |
| GIL limits true parallelism | Goroutines run in parallel across CPU cores |

Key ideas:

- **`defer`** schedules a call to run when the function returns, however it returns. It does the job of a `with` block's exit.
- **Concurrent map writes crash the program.** Go detects unsynchronised map access at runtime and stops with a fatal error. Python would not.
- **`sync.RWMutex`** allows many readers or one writer. It suits read-heavy APIs: use `RLock` for GETs and `Lock` for writes.
- **Grouping state in a struct** with its own lock and methods keeps locking in one place, rather than scattered across handlers.
- **Race detector:** `go test -race` and `go run -race .` report unsafe concurrent access. Use them.

The Python date search endpoint reads `posts` without the lock. Go's race detector would flag the same mistake.

---

## 9. Time and dates

Go's `time` package uses `time.Time` values and a layout system based on a fixed reference date.

```go
now := time.Now().UTC()
s := now.Format(time.RFC3339)               // "2026-10-07T19:30:00Z"
t, err := time.Parse(time.RFC3339, input)   // parse ISO 8601 input
if t.Before(end) && t.After(start) { ... }
```

| Python | Go |
|---|---|
| `datetime.utcnow()` | `time.Now().UTC()` |
| `.isoformat()` | `.Format(time.RFC3339)` |
| `datetime.fromisoformat(s)` | `time.Parse(time.RFC3339, s)` |
| `a <= b` | `!a.After(b)` |
| Naive vs aware datetimes | Every `time.Time` carries a location |

Points to know:

- **Store `time.Time`, not strings.** The JSON encoder writes `time.Time` as RFC 3339 automatically (with nanoseconds, e.g. `2026-10-07T19:46:46.731855898Z`), so you get correct output and easy comparisons. Call `.Truncate(time.Second)` when creating it if you want shorter timestamps.
- **No naive datetimes.** The Python version needed `.replace(tzinfo=...)` to avoid comparison errors, and that overwrote any supplied offset. Go parses the offset and compares instants correctly.
- **Layouts use a reference date**, `Mon Jan 2 15:04:05 MST 2006`, rather than `%Y-%m-%d` codes. The named constants such as `time.RFC3339` cover most needs.

---

## 10. Generating secure keys

Python's `secrets.token_urlsafe(20)` becomes `crypto/rand` plus URL-safe Base64.

```go
func newKey() (string, error) {
    b := make([]byte, 20)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(b), nil
}
```

Import `crypto/rand`, not `math/rand`. The `math` version is predictable and unsuitable for secrets.

To compare keys safely, `crypto/subtle.ConstantTimeCompare` avoids timing attacks. That is good practice, though not critical for a learning project.

---

## 11. Testing

Go has testing built in. Files ending in `_test.go` hold tests, and `go test` runs them.

```go
func TestCreateUser(t *testing.T) {
    req := httptest.NewRequest("POST", "/user",
        strings.NewReader(`{"user_screenname":"johndoe"}`))
    rec := httptest.NewRecorder()

    newMux().ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("got status %d, want 200", rec.Code)
    }
}
```

| Tool | Purpose |
|---|---|
| `testing` | Test functions, failures, subtests |
| `net/http/httptest` | Fake requests and recorded responses; no running server needed |
| Table-driven tests | One test function looping over many input and expected-output cases |
| `go test -race` | Tests with the race detector on |
| `go test -cover` | Coverage report |

**The existing Postman collection still works.** If the Go server listens on `127.0.0.1:5000` with the same routes and JSON, Newman can run the same collection against it. That makes the Postman tests a ready-made compatibility check between the two versions.

Go unit tests and Newman tests then complement each other: unit tests check handlers in isolation, and Newman checks the running API end to end.

---

## 12. Python to Go quick reference

| Concept | Python | Go |
|---|---|---|
| Variable | `x = 5` | `x := 5` or `var x int = 5` |
| Function | `def f(a):` | `func f(a int) int {` |
| Multiple returns | `return a, b` (tuple) | `return a, b` (built in) |
| None / null | `None` | `nil` (pointers, maps, slices, interfaces only) |
| List | `[]` | slice: `[]Post{}` |
| Append | `lst.append(x)` | `s = append(s, x)` |
| Dict | `{}` | `map[K]V{}` |
| String formatting | `f"id {x}"` | `fmt.Sprintf("id %d", x)` |
| String to int | `int(s)` | `strconv.Atoi(s)` |
| Class | `class User:` | `type User struct` plus methods |
| Method | `def save(self):` | `func (u *User) Save() {` |
| Exceptions | `try/except` | returned `error` values |
| Cleanup | `with lock:` | `defer mu.Unlock()` |
| Imports | `import json` | `import "encoding/json"` |
| Unused import or variable | Allowed | Compile error |
| Formatting | Optional (Black, etc.) | `go fmt` is standard |

---

## Further reading

- [A Tour of Go](https://go.dev/tour/): interactive introduction to the language
- [Effective Go](https://go.dev/doc/effective_go): idiomatic style and conventions
- [Go by Example](https://gobyexample.com/): short, runnable examples by topic
- [net/http package docs](https://pkg.go.dev/net/http): routing patterns and handlers
- [Routing enhancements for Go 1.22](https://go.dev/blog/routing-enhancements): the method and path-parameter syntax used above
