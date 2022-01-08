package main

import "net/http"

// SessionLoad keep track of sessions
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}
