package main

import "fmt"

type UserRepository struct{}

func (r *UserRepository) Save(name string) {
	fmt.Printf("Saving user %s to database...\n", name)
}
