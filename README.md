# Go_REST_WebForum

A RESTful backend for a simple web forum, being ported from **Python (Flask)** to **Go** as a learning exercise.

The original version was built in December 2023 as a project for a Python fundamentals course. This repository keeps that Flask implementation, its history and its Postman tests intact as a reference, and uses it as the specification for a Go rewrite.

---

## Why this project

- **Learn Go through a familiar problem.** The API's behaviour is already defined, so the focus stays on Go itself: types, HTTP handling, JSON, concurrency and error handling.
- **Compare two languages side by side.** The same endpoints, data model and tests, implemented twice.
- **Revisit earlier work.** The Python version has known issues. The port is a chance to understand why they happen and decide how Go should handle them.

---

## Status

| Version | State |
|---|---|
| Python / Flask | Complete (original coursework, unchanged) |
| Go | In progress |

---

## The API

The forum supports user accounts, posts, threaded replies and search. Data is held in memory and protected by a lock for concurrent access.

| Method | Endpoint | Purpose |
|---|---|---|
| `POST` | `/user` | Create a user with a unique screen name and secret key |
| `GET` | `/user` | List all users |
| `GET` | `/user/<id or screenname>` | Get a user's profile |
| `PUT` | `/user/<id>/edit` | Edit a user's profile |
| `POST` | `/post` | Create a post or a reply (via `post_parentid`) |
| `GET` | `/post/<id>` | Get a post and its direct replies |
| `DELETE` | `/post/<id>/delete/<key>` | Delete a post with its key and user credentials |
| `GET` | `/posts/search/datetime` | Find posts within an ISO 8601 date range |
| `GET` | `/posts/search/user` | Find all posts by a user |

### Extensions in the original coursework

1. **Users and user keys:** user creation with unique screen names.
2. **User profiles:** retrieve and edit user metadata.
3. **Threaded replies:** parent and child post IDs.
4. **Date/time search:** query posts by ISO 8601 range.
5. **User search:** return posts by user ID.

---

## File layout

```
.
├── app.py                              # Flask application and API endpoints
├── run.sh                              # Starts the Flask development server
├── setup.sh                            # Installs dependencies
├── test.sh                             # Runs Newman API tests
├── cdurkin_cs515_project3_collection.postman_collection.json
├── cdurkin_project3_enviroment.postman_environment.json
├── testrun_results/                    # Postman test run from the original submission
└── README.md
```

The Go implementation will be added alongside the Python files as it develops.

---

## Running the Python version

Requires Python 3 and Flask. Tests require [Newman](https://github.com/postmanlabs/newman).

```bash
./setup.sh   # install dependencies
./run.sh     # start the Flask server on http://127.0.0.1:5000
./test.sh    # run the Postman collection via Newman
```

---

## Known issues in the original

Recorded at submission and left unfixed in the Python code:

- **Delete endpoint:** returns HTTP 415 in some environments due to media type handling.
- **Reply parent check:** the reply test sometimes misaligns environment variables.
- **Autograder:** `test.sh` failed in CI due to an environment configuration mismatch.

The original Postman run passed 36 of 40 tests.
