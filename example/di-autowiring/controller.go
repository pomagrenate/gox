package main

import (
	"fmt"
	"gox/pkg/core"
)

type UserController struct {
	Service *UserService `inject:""`
}

type CreateRequest struct {
	Name string `json:"name" validate:"required"`
}

func (c *UserController) Create(ctx *core.Context, req *CreateRequest) (*core.Empty, error) {
	fmt.Printf("Received CreateUser request for %s\n", req.Name)
	c.Service.CreateUser(req.Name)
	return nil, nil
}
