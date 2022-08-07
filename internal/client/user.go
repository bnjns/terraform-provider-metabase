package client

import "fmt"

type User struct {
	Id         int64   `json:"id"`
	Email      string  `json:"email"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	CommonName string  `json:"common_name"`
	Groups     []int64 `json:"group_ids"`
	Locale     *string `json:"locale"`

	GoogleAuth bool `json:"google_auth"`
	LdapAuth   bool `json:"ldap_auth"`

	IsActive                bool `json:"is_active"`
	IsInstaller             bool `json:"is_installer"`
	IsQbnewb                bool `json:"is_qbnewb"`
	IsSuperuser             bool `json:"is_superuser"`
	HasInvitedSecondUser    bool `json:"has_invited_second_user"`
	HasQuestionAndDashboard bool `json:"has_question_and_dashboard"`

	DateJoined string  `json:"date_joined"`
	FirstLogin *string `json:"first_login"`
	LastLogin  *string `json:"last_login"`
	UpdatedAt  *string `json:"updated_at"`
}

func (c *Client) GetCurrentUser() (*User, error) {
	var currentUser User
	err := c.doGet("/user/current", &currentUser)
	if err != nil {
		return nil, err
	}

	return &currentUser, nil
}

func (c *Client) GetUser(userId int64) (*User, error) {
	var user User
	err := c.doGet(fmt.Sprintf("/user/%d", userId), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
