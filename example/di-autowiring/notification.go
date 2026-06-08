package main

import (
	"fmt"
	"gox/pkg/bus"
)

type NotificationService struct {
	Bus bus.EventBus `inject:""`
}

func (n *NotificationService) Init() {
	n.Bus.Subscribe("user.created", func(data any) {
		name := data.(string)
		fmt.Printf("[NotificationService] Sending welcome email to %s asynchronously...\n", name)
	})
}
