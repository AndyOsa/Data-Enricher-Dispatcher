package client

import (
  "context"
  "encoding/json"
  "net/http"

  "data-enricher-dispatcher/model"
)

type APIClient struct {
  apiA string
}

func NewAPIClient(apiA, _ string) *APIClient {
  return &APIClient{apiA: apiA}
}

func (c *APIClient) GetUsers(ctx context.Context) ([]model.User, error) {
  req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiA, nil)
  if err != nil {
    return nil, err
  }

  resp, err := http.DefaultClient.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()

  var users []model.User
  if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
    return nil, err
  }

  return users, nil
}
