# Go deployer Hook

[![Go Report Card](https://goreportcard.com/badge/github.com/wayand/deployer)](https://goreportcard.com/report/github.com/wayand/deployer)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

Go Deploy Hook is a simple and lightweight webhook server written in Go that allows you to automatically deploy your projects whenever a push event occurs on GitHub. It listens for GitHub webhooks and triggers a deployment script for each project. This server can manage multiple projects, each with its own endpoint, using a common deployment pattern.

### Features

- **Multiple Projects Support**: Each project has its own webhook endpoint, like `/ProjectName`.
- **Automatic Deployment**: When a push to the `master` branch is detected, the project is automatically pulled and rebuilt using Docker Compose.
- **Environment Configurable**: Uses environment variables (with support from `.env` files) to configure project paths.
- **Runs in Docker**: The application runs as a Docker container, making deployment easy and portable.

---

## How It Works

- When a push event is triggered on a GitHub repository, GitHub sends a webhook to the respective project endpoint, e.g., `https://deployer.mydomain.com/ProjectA`.
- The webhook server verifies the event, pulls the latest changes from the `master` branch, and redeploys the project using Docker Compose.

---

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- A GitHub repository set up to trigger webhooks.

### Installation

1. **Clone the repository:**

```bash
   git clone https://github.com/wayand/deployer.git
   cd deployer
```

2. **Configure the environment:**

```bash
   PROJECTS_FOLDER=/absolute/path/to/your/projects
```
   This environment variable defines the base directory where all your projects are stored.

3. **Set up Docker Compose:**
    The docker-compose.yml file is already configured. Simply build and run the Docker container:

```bash
    docker-compose up -d --build
```

4. **Expose the Webhook Service:**
    Make sure port 9000 is open on your firewall or reverse proxy so GitHub can reach your webhook server.

---
### Usage

1. **Configure Webhooks in GitHub:**

For each project, go to your GitHub repository and navigate to Settings > Webhooks > Add Webhook.
Set the Payload URL to the following format: https://deployer.mydomain.com/ProjectA (replace ProjectA with your project name).
Choose application/json as the content type.
Select Just the push event.
2. **Trigger Deployment:**

Once the webhook is configured, every time you push to the master branch of the respective project, the webhook server will automatically:

Pull the latest changes.
Stop and rebuild the project using Docker Compose.

---
### API Endpoints
- GET / – Root endpoint (returns a simple message).
- POST /ProjectA – Webhook endpoint for ProjectA (similarly for other projects).

---
### Environment Variables
- PROJECTS_FOLDER: The base directory where all projects are stored. This is required for deployment.

---
### Example Deployment Flow
Here is how the deployment works for a project named ProjectA:

- A webhook is triggered on https://deployer.mydomain.com/ProjectA when a push is made to the master branch of the corresponding GitHub repository.
- The webhook server pulls the latest changes from the repository.
- The server stops the currently running Docker container and rebuilds the project using docker-compose up -d --build.

---
### Project Structure

```text
    go-deploy-hook/
    │
    ├── .env                # Environment file for configuration
    ├── Dockerfile          # Dockerfile to build the Go webhook app
    ├── docker-compose.yml  # Docker Compose config to run the service
    ├── webhook.go          # Main Go application
    └── README.md           # Project documentation

```

---
### Development

1. Install dependencies:
```bash
    go mod tidy
```

2. Run the application locally (without Docker):
```bash
    go run main.go
```

3. Build the Go binary:
```bash
    go build -o webhook .
```

---
## Contributing
Contributions are welcome! Please feel free to submit a pull request or open an issue.

---
## License
This project is licensed under the MIT License.


### Key Sections Explained:
- **Overview**: Provides a concise description of the project.
- **How It Works**: Explains the workflow of the webhook server.
- **Getting Started**: Detailed instructions on installation, configuration, and running the project.
- **Usage**: Explains how to configure GitHub webhooks and use the app.
- **API Endpoints**: Lists the API endpoints available.
- **Environment Variables**: Describes how to configure the project path using a `.env` file.
- **Example Deployment Flow**: Shows how the deployment process works for each project.
- **Development**: Instructions for local development and building the Go app.
- **Contributing**: Encourages contributions from the community.
- **License**: States the project's MIT License.

This README should give users a good idea of what your project does, how to set it up, and how to use it. Let me know if you'd like to tweak any part of it! 😊
