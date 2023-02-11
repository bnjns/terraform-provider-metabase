package client

import (
	"fmt"
	"strings"
)

type GroupMembership struct {
	Id int64 `json:"id"`
}

type User struct {
	Id         int64   `json:"id"`
	Email      string  `json:"email"`
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	CommonName *string `json:"common_name"`
	Locale     *string `json:"locale"`

	GroupMemberships []GroupMembership `json:"user_group_memberships"`

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

type CurrentUser struct {
	User

	GroupIds []int64 `json:"group_ids"`
}

type CreateUserRequest struct {
	Email            string             `json:"email"`
	FirstName        *string            `json:"first_name"`
	LastName         *string            `json:"last_name"`
	GroupMemberships *[]GroupMembership `json:"user_group_memberships"`
}

type UpdateUserRequest struct {
	Email            *string            `json:"email"`
	FirstName        *string            `json:"first_name"`
	LastName         *string            `json:"last_name"`
	GroupMemberships *[]GroupMembership `json:"user_group_memberships"`
	Locale           *string            `json:"locale"`
	IsSuperuser      *bool              `json:"is_superuser"`
}

func (c *Client) GetCurrentUser() (*User, error) {
	var currentUser CurrentUser
	err := c.doGet("/user/current", &currentUser)
	if err != nil {
		return nil, err
	}

	groupMemberships := []GroupMembership{}
	if currentUser.GroupIds != nil {
		for _, groupId := range currentUser.GroupIds {
			groupMemberships = append(groupMemberships, GroupMembership{Id: groupId})
		}
	}

	user := User{
		Id:         currentUser.Id,
		Email:      currentUser.Email,
		FirstName:  currentUser.FirstName,
		LastName:   currentUser.LastName,
		CommonName: currentUser.CommonName,
		Locale:     currentUser.Locale,

		GroupMemberships: groupMemberships,

		GoogleAuth: currentUser.GoogleAuth,
		LdapAuth:   currentUser.LdapAuth,

		IsActive:                currentUser.IsActive,
		IsInstaller:             currentUser.IsInstaller,
		IsQbnewb:                currentUser.IsQbnewb,
		IsSuperuser:             currentUser.IsSuperuser,
		HasInvitedSecondUser:    currentUser.HasInvitedSecondUser,
		HasQuestionAndDashboard: currentUser.HasQuestionAndDashboard,

		DateJoined: currentUser.DateJoined,
		FirstLogin: currentUser.FirstLogin,
		LastLogin:  currentUser.LastLogin,
		UpdatedAt:  currentUser.UpdatedAt,
	}

	return &user, nil
}

func (c *Client) GetUser(userId int64) (*User, error) {
	var user User
	err := c.doGet(fmt.Sprintf("/user/%d", userId), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *Client) CreateUser(req CreateUserRequest) (int64, error) {
	var resp User
	err := c.doPost("/user", req, &resp)
	if err != nil {
		return 0, err
	}

	return resp.Id, nil
}

func (c *Client) UpdateUser(userId int64, req UpdateUserRequest) error {
	var resp User
	err := c.doPut(fmt.Sprintf("/user/%d", userId), req, &resp)
	return err
}

func (c *Client) ReactivateUser(userId int64) error {
	var resp User
	err := c.doPut(fmt.Sprintf("/user/%d/reactivate", userId), nil, &resp)
	if err != nil {
		if err.Error() == "Not found." {
			return ErrNotFound
		} else if strings.Contains(err.Error(), "Not able to reactivate an active user") {
			return nil
		} else {
			return err
		}
	}

	if !resp.IsActive {
		return fmt.Errorf("user was not updated to be active")
	}

	return nil
}

func (c *Client) DeleteUser(userId int64) error {
	var resp SuccessResponse
	err := c.doDelete(fmt.Sprintf("/user/%d", userId), &resp)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("API returned an unsuccessful response. Check the API logs for more details")
	}

	return nil
}
