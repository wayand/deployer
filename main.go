package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

// Payload structure for GitHub push events
type Payload struct {
	Ref string `json:"ref"`
}

func deployHandler(project string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload Payload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			fmt.Printf("Bad request for project %s: %v", project, err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Ensure we're only pulling if it's a push to the master branch
		if payload.Ref == "refs/heads/master" {
			projectsFolder := os.Getenv("PROJECTS_FOLDER")
			if projectsFolder == "" {
				fmt.Println("PROJECTS_FOLDER is not set")
				http.Error(w, "PROJECTS_FOLDER is not set", http.StatusInternalServerError)
				return
			}

			// Define the folder based on the project and the environment variable
			projectPath := fmt.Sprintf("%s/%s", projectsFolder, project)

			// Mark the project directory as safe for Git operations
			exec.Command("git", "config", "--global", "--add", "safe.directory", projectPath).Run()

			// Command to pull the latest code and redeploy using Docker Compose or any other method
			cmdText := fmt.Sprintf(`
                cd %s &&
				sudo chown -R root:root . &&
				sudo chmod -R 755 . &&
				git reset --hard &&
				git clean -fd &&
                git pull origin master &&
                docker-compose down &&
                docker-compose up -d --build
            `, projectPath)
			cmd := exec.Command("bash", "-c", cmdText)

			// Execute the deployment scripts
			if err := cmd.Run(); err != nil {
				fmt.Println("Error deploying: ", project, err)
				fmt.Println("we try to run this command: ", cmdText)
				http.Error(w, fmt.Sprintf("Error deploying %s: %v", project, err), http.StatusInternalServerError)
				return
			}

			fmt.Printf("Deployment for %s successful", project)
			fmt.Fprintf(w, "Deployment for %s successful", project)
		} else {
			fmt.Println("Not a master branch push")
			fmt.Fprintf(w, "Not a master branch push")
		}
	}
}

// Root handler to serve the root endpoint
func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("The root endpoint doesn't contain anything, please try: /ai-picture or /budget")
	fmt.Fprintf(w, "The root of Project, nothing here")
}

func main() {

	// Define the endpoints for each project
	http.HandleFunc("/ai-picture", deployHandler("ai-picture"))
	http.HandleFunc("/budget", deployHandler("budget"))
	http.HandleFunc("/invoicer", deployHandler("invoicer"))
	http.HandleFunc("/", rootHandler)

	// Listen on port 9000
	fmt.Println("Listening on :9000")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		fmt.Println("Error:", err)
	}
}
