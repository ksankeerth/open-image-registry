package mgmt

import "time"

type UserAccountViewDTO struct {
	Id                     string     `json:"id"`
	Username               string     `json:"username"`
	Email                  string     `json:"email"`
	DisplayName            string     `json:"display_name"`
	Locked                 bool       `json:"locked"`
	LockedReason           string     `json:"locked_reason"`
	LockedAt               *time.Time `json:"locked_at"`
	FailedAttempts         int        `json:"failed_attempts"`
	PasswordRecoveryId     string     `json:"password_recovery_id"`
	PasswordRecoveryReason string     `json:"password_recovery_reason"`
	PasswordRecoveryAt     *time.Time `json:"password_recovery_at"`
	LastLoggedInAt         *time.Time `json:"last_loggedin_at"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              *time.Time `json:"updated_at"`
	Role                   string     `json:"role"`
}

type CreateUserAccountRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type CreateUserAccountResponse struct {
	Username string `json:"username"`
	UserId   string `json:"user_id"`
}

type UpdateUserAccountRequest struct {
	DisplayName string `json:"display_name"`
}

type ChangeUserRoleRequest struct {
	Role string `json:"role"`
}

type PasswordChangeRequest struct {
	UserId      string `json:"user_id"`
	RecoveryId  string `json:"recovery_id"`
	Password    string `json:"password"`
	OldPassword string `json:"old_password"`
}

type ChangePasswordResponse struct {
	Status  string `json:"status"` // "success" or "failed"
	Message string `json:"message"`
}

type UpdateUserEmailRequest struct {
	UserId string `json:"user_id"`
	Email  string `json:"email"`
}

type UsernameEmailValidationRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type UsernameEmailValidationResponse struct {
	UsernameAvailable bool `json:"username_available"`
	EmailAvailable    bool `json:"email_available"`
}

type UserAccountSetupInfoResponse struct {
	ErrorMessage string `json:"error_message"`
	Username     string `json:"username"`
	UserId       string `json:"user_id"`
	DisplayName  string `json:"display_name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
}

type PasswordValidationRequest struct {
	Password string `json:"password"`
}

type AccountSetupCompleteRequest struct {
	UserId      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
	Uuid        string `json:"uuid"`
}

type GetUserResponse struct {
	Id             string     `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	DisplayName    string     `json:"display_name"`
	Locked         bool       `json:"locked"`
	LockedReason   int        `json:"locked_reason"`
	FailedAttempts int        `json:"failed_attempts"`
	Role           string     `json:"role"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
	LockedAt       *time.Time `json:"locked_at"`
}