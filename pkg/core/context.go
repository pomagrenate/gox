package core

import (
	"encoding/json"
	"net/http"
)

// Context wraps http.ResponseWriter and http.Request to provide convenient methods.
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

// NewContext creates a new Context.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
	}
}

// JSON sends a JSON response with the specified status code.
func (c *Context) JSON(code int, obj any) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(code)
	if err := json.NewEncoder(c.Writer).Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// Param returns the path parameter by name.
// Relies on Go 1.22+ wildcard routing (e.g., /users/{id})
func (c *Context) Param(name string) string {
	return c.Request.PathValue(name)
}

// Query returns the query parameter by name.
func (c *Context) Query(name string) string {
	return c.Request.URL.Query().Get(name)
}
