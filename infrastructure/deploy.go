package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createServiceDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

func deployStack(repo, branch, env, webImage string) error {
	// Configuration
	stackName := fmt.Sprintf("%s_%s_%s", repo, branch, env)
	serviceDir := fmt.Sprintf("%s_%s_%s", repo, strings.ReplaceAll(branch, "/", "-"), env)
	dbServiceDir := serviceDir + "_db"
	webServiceDir := serviceDir + "_web"

	// Create service directories
	err := createServiceDirectory(filepath.Join("/swarm01-data/services", dbServiceDir))
	if err != nil {
		return fmt.Errorf("failed to create DB service directory: %v", err)
	}
	err = createServiceDirectory(filepath.Join("/swarm01-data/services", webServiceDir))
	if err != nil {
		return fmt.Errorf("failed to create Web service directory: %v", err)
	}

	// Set environment variables for docker-compose
	os.Setenv("DB_PASSWORD", "YourStrong@Passw0rd") // Consider using a more secure method to set this
	os.Setenv("DB_PORT", "1433")                    // You might want to make this dynamic
	os.Setenv("WEB_PORT", "8080")                   // You might want to make this dynamic
	os.Setenv("SERVICE_DIR", dbServiceDir)
	os.Setenv("WEB_IMAGE", webImage)

	// Deploy the stack
	err = runCommand("docker", "stack", "deploy", "-c", "docker-compose.yml", stackName)
	if err != nil {
		return fmt.Errorf("failed to deploy stack: %v", err)
	}

	fmt.Printf("Stack '%s' deployed successfully\n", stackName)
	return nil
}

func main() {
	if len(os.Args) != 5 {
		log.Fatalf("Usage: %s <repo> <branch> <environment> <web_image>", os.Args[0])
	}

	repo := os.Args[1]
	branch := os.Args[2]
	env := os.Args[3]
	webImage := os.Args[4]

	err := deployStack(repo, branch, env, webImage)
	if err != nil {
		log.Fatalf("Error deploying stack: %v", err)
	}
}
