package main

import "gox/pkg/bus"

type UserService struct {
	Repo *UserRepository `inject:""`
	Bus  bus.EventBus    `inject:""`
}

func (s *UserService) CreateUser(name string) {
	s.Repo.Save(name)
	// Publish an event that the user was created
	s.Bus.Publish("user.created", name)
}
