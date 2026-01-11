package mgmt

type UserProfileInfo struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type AuthLoginResponse struct {
	User UserProfileInfo `json:"user"`
	// TODO: later, We may send additional information such as resources the user has access 
}

type AuthLoginRequest struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Scopes   []string `json:"scopes"`
}