package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
)

// Payload structure for GitHub push events
type Payload struct {
	Ref string `json:"ref"`
}


// Load environment variables
func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func deployHandler(project string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload Payload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Ensure we're only pulling if it's a push to the master branch
		if payload.Ref == "refs/heads/master" {
			projectsFolder := os.Getenv("PROJECTS_FOLDER")
			if projectsFolder == "" {
				http.Error(w, "PROJECTS_FOLDER is not set", http.StatusInternalServerError)
				return
			}

			// Define the folder based on the project and the environment variable
			projectPath := fmt.Sprintf("%s/%s", projectsFolder, project)
			// Command to pull the latest code and redeploy using Docker Compose or any other method
			cmd := exec.Command("bash", "-c", fmt.Sprintf(`
                cd %s &&
                git pull origin master &&
                docker-compose down &&
                docker-compose up -d --build
            `, projectPath))

			// Execute the deployment script
			if err := cmd.Run(); err != nil {
				http.Error(w, fmt.Sprintf("Error deploying %s: %v", project, err), http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "Deployment for %s successful", project)
		} else {
			fmt.Fprintf(w, "Not a master branch push")
		}
		fmt.Println("payload: ", payload)
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
