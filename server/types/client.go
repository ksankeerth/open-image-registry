package types

import "time"

type DockerRegistryLoginResponse struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int64     `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}