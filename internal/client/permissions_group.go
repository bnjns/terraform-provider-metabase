package client

import "fmt"

type PermissionsGroup struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type PermissionsGroupRequest struct {
	Name string `json:"name"`
}

func (c *Client) GetPermissionsGroup(groupId int64) (*PermissionsGroup, error) {
	var group PermissionsGroup
	err := c.doGet(fmt.Sprintf("/permissions/group/%d", groupId), &group)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (c *Client) CreatePermissionsGroup(req PermissionsGroupRequest) (int64, error) {
	var resp PermissionsGroup
	err := c.doPost("/permissions/group", req, &resp)
	if err != nil {
		return 0, nil
	}

	return resp.Id, nil
}

func (c *Client) UpdatePermissionsGroup(groupId int64, req PermissionsGroupRequest) error {
	var resp PermissionsGroup
	err := c.doPut(fmt.Sprintf("/permissions/group/%d", groupId), req, &resp)
	return err
}

func (c *Client) DeletePermissionsGroup(groupId int64) error {
	err := c.doDelete(fmt.Sprintf("/permissions/group/%d", groupId), nil)
	if err != nil {
		return err
	}

	return nil
}
