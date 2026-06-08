package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"gox/pkg/core"
)

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `query:"age" validate:"gte=18"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type UserController struct{}

// Create handles the POST /users request.
// Notice there is ZERO boilerplate: no http.ResponseWriter, no body decoding, no validation calls.
func (c *UserController) Create(ctx *core.Context, req *CreateUserRequest) (*UserResponse, error) {
	// If we reach here, `req` is already validated and populated!
	fmt.Printf("Received valid request: %+v\n", req)

	// Simulate business logic
	return &UserResponse{
		ID:    "usr_123456",
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}, nil
}

type GetUserRequest struct {
	ID string `path:"id" validate:"required,min=5"`
}

// Get handles the GET /users/{id} request.
func (c *UserController) Get(ctx *core.Context, req *GetUserRequest) (*UserResponse, error) {
	// If we reach here, `req.ID` is bound from the URL path and validated.
	fmt.Printf("Fetching user ID: %s\n", req.ID)

	return &UserResponse{
		ID:    req.ID,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}, nil
}

func main() {
	ctrl := &UserController{}

	mux := http.NewServeMux()

	// Bind GoX Generic Handlers
	mux.HandleFunc("POST /users", core.Handler(ctrl.Create))
	mux.HandleFunc("GET /users/{id}", core.Handler(ctrl.Get))

	// Add a slow endpoint to simulate long-running tasks for Zero-Downtime Live Reload tests
	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handling /slow request...")
		time.Sleep(5 * time.Second)
		w.Write([]byte("Hello GoX V2 - slow response finished!"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		fmt.Printf("Starting zero-boilerplate server on port %s...\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
	fmt.Println("Server exiting")
}
