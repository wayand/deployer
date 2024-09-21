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

// Function to handle the deployment in the background
func deploy(project string, projectPath string) {

	// Mark the project directory as safe for Git operations
	// exec.Command("git", "config", "--global", "--add", "safe.directory", projectPath).Run()

	token := os.Getenv("GITHUB_TOKEN")
	user := os.Getenv("GITHUB_USER")
	if token == "" || user == "" {
		fmt.Println("token or user is not set")
		return
	}

	// GitHub repo URL using HTTPS and the token for authentication
	repoURL := fmt.Sprintf("https://%s@github.com/%s/%s.git", token, user, project)

	// Command to pull the latest code and redeploy using Docker Compose
	cmdText := fmt.Sprintf(`
                cd %s &&
				git remote set-url origin %s &&
                git fetch --depth=1 origin master &&
				git reset --hard FETCH_HEAD &&
                docker-compose down &&
                docker-compose up -d --build
            `, projectPath, repoURL)
	cmd := exec.Command("bash", "-c", cmdText)

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprintf("Error deploying %s: %v\n%s", project, err, string(output)))
		fmt.Println("we try to run this command: ", cmdText)
	} else {
		fmt.Printf("Deployment for %s successful\n%s", project, string(output))
	}
}

func deployHandler(project string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload Payload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			fmt.Println(fmt.Sprintf("Bad request for project %s: %v", project, err))
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

			// Respond immediately to the webhook request to avoid GitHub timeout
			fmt.Fprintf(w, "Webhook received. Deployment started for %s", project)

			// Run the deployment process in the background
			go deploy(project, projectPath)

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
