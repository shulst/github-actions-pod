package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type GitHubEvent struct {
	Ref        string `json:"ref"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run test_script.go <repository_name> <branch_name>")
		os.Exit(1)
	}

	repoName := os.Args[1]
	branchName := os.Args[2]

	event := GitHubEvent{
		Ref: fmt.Sprintf("refs/heads/%s", branchName),
		Repository: struct {
			Name string `json:"name"`
		}{
			Name: repoName,
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post("http://localhost:8080/webhook", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response status: %s\n", resp.Status)
	fmt.Printf("Response body: %s\n", string(body))
}
