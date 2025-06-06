package service

import (
  "bytes"
  "context"
  "encoding/json"
  "log"
  "net/http"
  "strings"
  "time"

  "data-enricher-dispatcher/model"
)

type Dispatcher struct {
  apiB string
}

func NewDispatcher(_ string, apiB string) *Dispatcher {
  return &Dispatcher{apiB: apiB}
}

func (d *Dispatcher) ProcessUsers(ctx context.Context, users []model.User) {
  for _, user := range users {
    if strings.HasSuffix(user.Email, ".biz") {
      err := d.sendWithRetry(ctx, user, 3)
      if err != nil {
        log.Printf("Failed to send user %s: %v", user.Email, err)
      }
    } else {
      log.Printf("Skipping user %s (email does not end with .biz)", user.Email)
    }
  }
}

func (d *Dispatcher) sendWithRetry(ctx context.Context, user model.User, retries int) error {
  body, _ := json.Marshal(map[string]string{
    "name":  user.Name,
    "email": user.Email,
  })

  for i := 0; i < retries; i++ {
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.apiB, bytes.NewBuffer(body))
    if err != nil {
      return err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
      log.Printf("Successfully sent user: %s", user.Email)
      return nil
    }

    if resp != nil {
      log.Printf("Attempt %d: failed with status %d", i+1, resp.StatusCode)
      resp.Body.Close()
    } else {
      log.Printf("Attempt %d: request error: %v", i+1, err)
    }

    time.Sleep(2 * time.Second)
  }

  return context.DeadlineExceeded
}
