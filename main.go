package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/rs/zerolog/log"
)

func main() {
	var loginAt time.Time
	http.HandleFunc("GET /content", http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			if loginAt.IsZero() || time.Since(loginAt) > 1*time.Second {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			fmt.Println("get content!")
		}))
	http.HandleFunc("GET /login", http.HandlerFunc(
		func(_ http.ResponseWriter, _ *http.Request) {
			loginAt = time.Now()

			fmt.Println("login!")
		}))
	go http.ListenAndServe(addr, nil)

	time.Sleep(1 * time.Second)

	client := &http.Client{
		Transport: new(rt),
	}

	for range 11 {
		resp, err := client.Get(contentURL)
		if err != nil {
			log.Fatal().Err(err).Msg("make request")
		}
		c, _ := httputil.DumpResponse(resp, true)
		fmt.Println(">> resp:")
		fmt.Print(string(c))

		time.Sleep(100 * time.Millisecond)
	}
}

type rt int

func (*rt) RoundTrip(r *http.Request) (*http.Response, error) {
	for range 3 {
		resp, err := http.DefaultClient.Do(r)
		if err != nil || resp.StatusCode != http.StatusUnauthorized {
			return resp, err
		}
		req, err := http.NewRequest(http.MethodGet, loginURL, nil)
		if err != nil {
			return nil, err
		}
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("status: %s", resp.Status)
		}
	}
	return nil, nil
}

const (
	addr       = "127.0.0.1:8080"
	loginURL   = "http://" + addr + "/login"
	contentURL = "http://" + addr + "/content"
)
