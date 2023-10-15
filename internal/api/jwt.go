package api

import "github.com/golang-jwt/jwt/v5"

var JWTKey = []byte("mttAG8HhKpRROKpsQ9dX7vZGhNnbRg8S")

type UserClaims struct {
	jwt.RegisteredClaims
	Uid          int64
	RefreshCount int64
	UserAgent    string
}
