package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type BranchConfig struct {
	Environment string `yaml:"environment"`
	Domain      string `yaml:"domain"`
}

type RepoConfig struct {
	Branches map[string]BranchConfig `yaml:"branches"`
}

type Config struct {
	Routes []Route `yaml:"routes"`
}

type Route struct {
	Repo        string `yaml:"repo"`
	Branch      string `yaml:"branch"`
	Environment string `yaml:"environment"`
	Domain      string `yaml:"domain"`
	Service     string `yaml:"service"`
}

var config Config
var repoConfigs map[string]RepoConfig

func loadConfig() {
	data, err := ioutil.ReadFile("./config/edge-config.yaml")
	if err != nil {
		log.Printf("Error reading config file: %v", err)
		config = Config{Routes: []Route{}}
		return
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}
}

func loadRepoConfigs() {
	data, err := ioutil.ReadFile("./config/branch-environment-mapping.yaml")
	if err != nil {
		log.Fatalf("Error reading repo configs file: %v", err)
	}
	err = yaml.Unmarshal(data, &repoConfigs)
	if err != nil {
		log.Fatalf("Error parsing repo configs: %v", err)
	}
}

func saveConfig() {
	data, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatalf("Error marshaling config: %v", err)
	}
	err = ioutil.WriteFile("./config/edge-config.yaml", data, 0644)
	if err != nil {
		log.Fatalf("Error writing config file: %v", err)
	}
}

func getBranchConfig(repo, branch string) (BranchConfig, bool) {
	repoConfig, ok := repoConfigs[repo]
	if !ok {
		return BranchConfig{}, false
	}

	// Check for exact match
	if branchConfig, ok := repoConfig.Branches[branch]; ok {
		return branchConfig, true
	}

	// Check for wildcard match
	for pattern, branchConfig := range repoConfig.Branches {
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(branch, prefix) {
				return branchConfig, true
			}
		}
	}

	return BranchConfig{}, false
}

func sanitizeBranchName(branch string) string {
	// Replace '/' with '-'
	branch = strings.ReplaceAll(branch, "/", "-")

	// Remove any character that's not a letter, number, or hyphen
	reg, err := regexp.Compile("[^a-zA-Z0-9-]+")
	if err != nil {
		log.Fatal(err)
	}
	branch = reg.ReplaceAllString(branch, "")

	// Ensure the branch name doesn't start or end with a hyphen
	branch = strings.Trim(branch, "-")

	// Lowercase the branch name
	branch = strings.ToLower(branch)

	// Truncate to a maximum of 63 characters (DNS label limit)
	if len(branch) > 63 {
		branch = branch[:63]
	}

	return branch
}

func logMessage(message string) {
	log.Println(message)
	fmt.Fprintln(os.Stdout, message) // This ensures the message is printed to the console
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logMessage(fmt.Sprintf("Error reading request body: %v", err))
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var event struct {
		Ref        string `json:"ref"`
		Repository struct {
			Name string `json:"name"`
		} `json:"repository"`
	}
	err = json.Unmarshal(payload, &event)
	if err != nil {
		logMessage(fmt.Sprintf("Error parsing JSON: %v", err))
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	repo := event.Repository.Name
	branch := event.Ref[11:] // Remove "refs/heads/" prefix
	sanitizedBranch := sanitizeBranchName(branch)

	_, repoFound := repoConfigs[repo]
	if !repoFound {
		errorMsg := fmt.Sprintf("Repository not found in configuration: %s", repo)
		logMessage(errorMsg)
		http.Error(w, errorMsg, http.StatusNotFound)
		return
	}

	branchConfig, branchFound := getBranchConfig(repo, branch)
	if !branchFound {
		logMessage(fmt.Sprintf("No specific configuration found for repo %s and branch %s. Using default configuration.", repo, branch))
		return
	}

	domain := strings.ReplaceAll(branchConfig.Domain, "{branch}", sanitizedBranch)

	// Update or add route
	found := false
	for i, route := range config.Routes {
		if route.Repo == repo && route.Branch == branch {
			config.Routes[i].Environment = branchConfig.Environment
			config.Routes[i].Domain = domain
			config.Routes[i].Service = fmt.Sprintf("%s_%s_%s_web", repo, sanitizedBranch, branchConfig.Environment)
			found = true
			break
		}
	}
	if !found {
		newRoute := Route{
			Repo:        repo,
			Branch:      branch,
			Environment: branchConfig.Environment,
			Domain:      domain,
			Service:     fmt.Sprintf("%s_%s_%s_web", repo, sanitizedBranch, branchConfig.Environment),
		}
		config.Routes = append(config.Routes, newRoute)
	}

	saveConfig()
	successMsg := fmt.Sprintf("Route updated for repo: %s, branch: %s, environment: %s, domain: %s", repo, branch, branchConfig.Environment, domain)
	logMessage(successMsg)
	fmt.Fprint(w, successMsg)
}

func main() {
	// Set up logging to write to both file and console
	logFile, err := os.OpenFile("webhook.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))

	loadConfig()
	loadRepoConfigs()
	http.HandleFunc("/webhook", webhookHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
