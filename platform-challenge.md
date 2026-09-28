# Platform challenge

## Objective

Build a simple Go echo web service, containerize it, and write Go Pulumi
infrastructure to deploy it to a local Kind Kubernetes cluster.

## Requirements

- Listen on a configurable port, defaulting to 8080, and handle any route.
- Return JSON containing request headers, query parameters, raw body, and path.
- Use Pulumi with local filesystem state, with no cloud backend required.
- Create a Kubernetes Deployment and Service.
- Provide a shell CI script that runs unit tests, builds a Docker image, and
  makes it available to Kind through a registry or `kind load`.
- Provide a GitHub Actions workflow executing the CI script on every push.
- Include a short SETUP.md covering installation and deployment from scratch.
- Submit a GitHub repository containing the source, Dockerfile, Pulumi project,
  CI script, workflow, and setup documentation.

Evaluation focuses on idiomatic code, clear infrastructure, simplicity, and
reproducible setup instructions.
