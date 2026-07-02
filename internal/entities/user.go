package entities

type Login struct {
	Token        string
	Expire       int64
	RefreshToken string
}
