package common

import "time"

const (
	JwtSalt = "itoio-jwt-salt"
	JwtExp  = time.Minute * 10
)
