package entity

import (
	"go-rest-skeleton/infrastructure/message/exception"
	"go-rest-skeleton/pkg/security"
	"html"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/google/uuid"
)

// User represent schema of table users.
type User struct {
	UUID       string      `gorm:"size:36;not null;unique_index;primary_key;" json:"uuid"`
	FirstName  string      `gorm:"size:100;not null;" json:"first_name"`
	LastName   string      `gorm:"size:100;not null;" json:"last_name"`
	Email      string      `gorm:"size:100;not null;unique;index:email" json:"email" form:"email"`
	Phone      string      `gorm:"size:100;" json:"phone,omitempty"`
	Password   string      `gorm:"size:100;not null;index:password" json:"password" form:"password"`
	AvatarUUID string      `gorm:"size:36;" json:"avatar_uuid"`
	CreatedAt  time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt  *time.Time  `json:"deleted_at,omitempty"`
	UserRoles  []UserRole  `gorm:"foreignKey:UserUUID"`
	UserLogins []UserLogin `gorm:"foreignKey:UserUUID"`
}

type UserResetPassword struct {
	NewPassword     string `json:"new_password" form:"new_password"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
}

// UserFaker represent content when generate fake data of user.
type UserFaker struct {
	UUID      string `faker:"uuid_hyphenated"`
	FirstName string `faker:"first_name"`
	LastName  string `faker:"last_name"`
	Email     string `faker:"email"`
	Phone     string `faker:"phone_number"`
	Password  string `faker:"password"`
}

// Users represent multiple User.
type Users []User

// DetailUser represent format of detail User.
type DetailUser struct {
	UserFieldsForDetail
	Role []interface{} `json:"roles,omitempty"`
}

// DetailUserList represent format of DetailUser for User list.
type DetailUserList struct {
	UserFieldsForDetail
	UserFieldsForList
}

// UserFieldsForDetail represent fields of detail User.
type UserFieldsForDetail struct {
	UUID      string      `json:"uuid"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Email     string      `json:"email"`
	Phone     interface{} `json:"phone,omitempty"`
	Avatar    interface{} `json:"avatar,omitempty"`
}

// UserFieldsForList represent fields of detail User for User list.
type UserFieldsForList struct {
	CreatedAt time.Time `json:"created_at"`
}

// Prepare will prepare submitted data of user.
func (u *User) Prepare() {
	u.FirstName = html.EscapeString(strings.TrimSpace(u.FirstName))
	u.LastName = html.EscapeString(strings.TrimSpace(u.LastName))
	u.Email = html.EscapeString(strings.TrimSpace(u.Email))
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
}

// BeforeSave handle uuid generation and password hashing.
func (u *User) BeforeSave() error {
	generateUUID := uuid.New()
	hashPassword, err := security.Hash(u.Password)
	if err != nil {
		return err
	}
	if u.UUID == "" {
		u.UUID = generateUUID.String()
	}
	u.Password = string(hashPassword)
	return nil
}

// DetailUsers will return formatted user detail of multiple user.
func (users Users) DetailUsers() []interface{} {
	result := make([]interface{}, len(users))
	for index, user := range users {
		result[index] = user.DetailUserList()
	}
	return result
}

// DetailUser will return formatted user detail of user.
func (u *User) DetailUser() interface{} {
	return &DetailUser{
		UserFieldsForDetail: UserFieldsForDetail{
			UUID:      u.UUID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Email:     u.Email,
			Phone:     u.Phone,
			Avatar:    nil,
		},
		Role: UserRoles.GetUserRole(u.UserRoles),
	}
}

// DetailUserAvatar will return formatted user detail of user.
func (u *User) DetailUserAvatar(url interface{}) interface{} {
	return &DetailUser{
		UserFieldsForDetail: UserFieldsForDetail{
			UUID:      u.UUID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Email:     u.Email,
			Phone:     u.Phone,
			Avatar:    url,
		},
		Role: UserRoles.GetUserRole(u.UserRoles),
	}
}

// DetailUserList will return formatted user detail of user for user list.
func (u *User) DetailUserList() interface{} {
	return &DetailUserList{
		UserFieldsForDetail: UserFieldsForDetail{
			UUID:      u.UUID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Email:     u.Email,
			Phone:     u.Phone,
		},
		UserFieldsForList: UserFieldsForList{
			CreatedAt: u.CreatedAt,
		},
	}
}

// ValidateSaveUser will validate create a new user request.
func (u *User) ValidateSaveUser() []exception.ErrorForm {
	var errMsg []exception.ErrorForm
	var errMsgData = make(map[string]interface{})
	var err error
	if u.FirstName == "" {
		errMsgData["Field"] = "first_name"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "first_name",
			Msg:   "api.msg.error.user.field_first_name_is_required",
			Data:  errMsgData})
	}
	if u.LastName == "" {
		errMsgData["Field"] = "last_name"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "last_name",
			Msg:   "api.msg.error.user.field_last_name_is_required",
			Data:  errMsgData})
	}
	if u.Password == "" {
		errMsgData["Field"] = "password"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "password",
			Msg:   "api.msg.error.user.field_password_is_required",
			Data:  errMsgData})
	}
	if u.Password != "" && len(u.Password) < 6 {
		errMsgData["Field"] = "password"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "password",
			Msg:   "api.msg.error.invalid_password_length",
			Data:  errMsgData})
	}
	if u.Email == "" {
		errMsgData["Field"] = "email"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "email",
			Msg:   "api.msg.error.user.field_email_is_required",
			Data:  errMsgData})
	}
	if u.Email != "" {
		if err = checkmail.ValidateFormat(u.Email); err != nil {
			errMsg = append(errMsg, exception.ErrorForm{
				Field: "email",
				Msg:   "api.msg.error.invalid_email"})
		}
	}
	return errMsg
}

// ValidateUpdateUser will validate create a new user request.
func (u *User) ValidateUpdateUser() []exception.ErrorForm {
	var errMsg []exception.ErrorForm
	var errMsgData = make(map[string]interface{})
	var err error
	if u.FirstName == "" {
		errMsgData["Field"] = "first_name"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "first_name",
			Msg:   "api.msg.error.user.field_first_name_is_required",
			Data:  errMsgData})
	}
	if u.LastName == "" {
		errMsgData["Field"] = "last_name"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "last_name",
			Msg:   "api.msg.error.user.field_last_name_is_required",
			Data:  errMsgData})
	}
	if u.Email == "" {
		errMsgData["Field"] = "email"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "email",
			Msg:   "api.msg.error.user.field_email_is_required",
			Data:  errMsgData})
	}
	if u.Email != "" {
		if err = checkmail.ValidateFormat(u.Email); err != nil {
			errMsg = append(errMsg, exception.ErrorForm{
				Field: "email",
				Msg:   "api.msg.error.invalid_email"})
		}
	}
	return errMsg
}

// ValidateLogin will validate login request.
func (u *User) ValidateLogin() []exception.ErrorForm {
	var errMsg []exception.ErrorForm
	var errMsgData = make(map[string]interface{})
	var err error
	if u.Password == "" {
		errMsgData["Field"] = "email"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "email",
			Msg:   "api.msg.error.user.field_email_is_required",
			Data:  errMsgData})
	}
	if u.Email == "" {
		errMsgData["Field"] = "password"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "password",
			Msg:   "api.msg.error.user.field_password_is_required",
			Data:  errMsgData})
	}
	if u.Email != "" {
		if err = checkmail.ValidateFormat(u.Email); err != nil {
			errMsg = append(errMsg, exception.ErrorForm{
				Field: "email",
				Msg:   "api.msg.error.invalid_email"})
		}
	}
	return errMsg
}

// ValidateForgotPassword will validate forgot password request.
func (u *User) ValidateForgotPassword() []exception.ErrorForm {
	var errMsg []exception.ErrorForm
	var errMsgData = make(map[string]interface{})
	var err error
	if u.Email == "" {
		errMsgData["Field"] = "email"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "email",
			Msg:   "api.msg.error.user.field_email_is_required",
			Data:  errMsgData})
	}
	if u.Email != "" {
		if err = checkmail.ValidateFormat(u.Email); err != nil {
			errMsg = append(errMsg, exception.ErrorForm{
				Field: "email",
				Msg:   "api.msg.error.invalid_email"})
		}
	}
	return errMsg
}

// ValidateResetPassword will validate reset password request.
func (u *UserResetPassword) ValidateResetPassword() []exception.ErrorForm {
	var errMsg []exception.ErrorForm
	var errMsgData = make(map[string]interface{})
	if u.NewPassword == "" {
		errMsgData["Field"] = "new_password"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "new_password",
			Msg:   "api.msg.error.user.field_new_password_is_required",
			Data:  errMsgData})
	}
	if u.ConfirmPassword == "" {
		errMsgData["Field"] = "confirm_password"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "confirm_password",
			Msg:   "api.msg.error.user.field_confirm_password_is_required",
			Data:  errMsgData})
	}
	if u.NewPassword != u.ConfirmPassword {
		errMsgData["Field"] = "new_password"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "new_password",
			Msg:   "api.msg.error.user.field_new_and_confirm_password_does_not_match",
			Data:  errMsgData})
		errMsgData["Field"] = "confirm_password"
		errMsg = append(errMsg, exception.ErrorForm{
			Field: "confirm_password",
			Msg:   "api.msg.error.user.field_new_and_confirm_password_does_not_match",
			Data:  errMsgData})
	}
	return errMsg
}
