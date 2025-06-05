// main.go
package main

import (
	"context"
	"log"
	"os"

	"data-enricher-dispatcher/client"
	"data-enricher-dispatcher/service"
)

func main() {
	ctx := context.Background()

	apiA := os.Getenv("API_A")
	apiB := os.Getenv("API_B")
	if apiA == "" {
		apiA = "https://jsonplaceholder.typicode.com/users"
	}
	if apiB == "" {
		log.Fatal("API_B must be set (e.g., webhook.site URL)")
	}

	c := client.NewAPIClient(apiA, apiB)
	d := service.NewDispatcher(c)
	if err := d.Process(ctx); err != nil {
		log.Fatalf("Processing failed: %v", err)
	}
}

// model/user.go
package model

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// client/api_client.go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"data-enricher-dispatcher/model"
)

type APIClient struct {
	sourceURL string
	targetURL string
	client    *http.Client
}

func NewAPIClient(source, target string) *APIClient {
	return &APIClient{
		sourceURL: source,
		targetURL: target,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *APIClient) FetchUsers(ctx context.Context) ([]model.User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.sourceURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch users")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var users []model.User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (a *APIClient) PostUser(ctx context.Context, user model.User) error {
	data, _ := json.Marshal(user)
	var err error
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, a.targetURL, bytes.NewBuffer(data))
		req.Header.Set("Content-Type", "application/json")
		resp, err := a.client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		log.Printf("POST failed (attempt %d): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	return err
}

// service/dispatcher.go
package service

import (
	"context"
	"fmt"
	"strings"

	"data-enricher-dispatcher/client"
	"data-enricher-dispatcher/model"
)

type Dispatcher struct {
	client *client.APIClient
}

func NewDispatcher(c *client.APIClient) *Dispatcher {
	return &Dispatcher{client: c}
}

func (d *Dispatcher) Process(ctx context.Context) error {
	users, err := d.client.FetchUsers(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		if strings.HasSuffix(user.Email, ".biz") {
			err := d.client.PostUser(ctx, user)
			if err != nil {
				fmt.Printf("Failed to post user %s: %v\n", user.Name, err)
			}
		} else {
			fmt.Printf("Skipping user: %s with email %s\n", user.Name, user.Email)
		}
	}
	return nil
}
