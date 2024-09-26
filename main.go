package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

// Payload structure for GitHub push events
type Payload struct {
	Ref string `json:"ref"`
}

// Function to verify the HMAC signature
func verifySignature(secret, signature string, body []byte) bool {
	// Create a new HMAC by defining the hash type and the secret key
	h := hmac.New(sha256.New, []byte(secret))

	// Write the body to it
	h.Write(body)

	// Get the final HMAC SHA256 code
	expectedMAC := hex.EncodeToString(h.Sum(nil))

	// Compare the provided signature with the calculated one
	// GitHub sends the signature as "sha256=<signature>"
	return hmac.Equal([]byte(signature[7:]), []byte(expectedMAC))
}

// Function to handle the deployment in the background
func deploy(project string, projectPath string) {

	// Mark the project directory as safe for Git operations
	exec.Command("git", "config", "--global", "--add", "safe.directory", projectPath).Run()

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
		// Read the secret from the environment variable
		secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
		if secret == "" {
			fmt.Println("Webhook secret not configured")
			http.Error(w, "Webhook secret not configured", http.StatusInternalServerError)
			return
		}

		// Get the signature from the request header
		signature := r.Header.Get("X-Hub-Signature-256")
		if signature == "" {
			fmt.Println("(X-Hub-Signature-256) Signature missing")
			http.Error(w, "Signature missing", http.StatusUnauthorized)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Could not read request body")
			http.Error(w, "Could not read request body", http.StatusInternalServerError)
			return
		}

		// Verify the HMAC signature
		if !verifySignature(secret, signature, body) {
			fmt.Println("Invalid signature")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		// If the signature is valid, proceed with the webhook processing
		var payload Payload
		if err := json.Unmarshal(body, &payload); err != nil {
			fmt.Println(fmt.Sprintf("Bad request for project %s: %v", project, err))
			http.Error(w, "Bad request", http.StatusBadRequest)
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
			fmt.Fprintf(w, "Webhook verified. Deployment started for %s", project)

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
	HTTP_PORT := os.Getenv("HTTP_PORT")
	fmt.Printf("Listening on %s", HTTP_PORT)
	if err := http.ListenAndServe(HTTP_PORT, nil); err != nil {
		fmt.Println("Error:", err)
	}
}
