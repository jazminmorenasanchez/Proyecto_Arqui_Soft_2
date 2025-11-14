package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type UsersClient struct {
	base string
	http *http.Client
}

type UserDTO struct {
	ID    uint64 `json:"id"`
	Role  string `json:"rol"`
	Email string `json:"email"`
}

func NewUsersClient(base string) *UsersClient {
	return &UsersClient{
		base: base,
		http: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *UsersClient) GetUser(id string) (*UserDTO, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/users/%s", c.base, id), nil)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("users-api returned %d", res.StatusCode)
	}
	var u UserDTO
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}
