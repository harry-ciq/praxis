//go:build tools

package tools

// This file ensures tool/future dependencies remain in go.mod.
// These will be imported by application code as features are built.
import (
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/google/go-github/v60/github"
	_ "github.com/gorilla/websocket"
)
