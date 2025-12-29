# GitHub Raw URL Explorer

[![GitHub Repo](https://img.shields.io/badge/GitHub-Repository-blue?logo=github)](https://github.com/florianibach/github-repo-raw-urls-ui)

[![DockerHub Repo](https://img.shields.io/badge/Docker_Hub-Repository-blue?logo=docker)](https://hub.docker.com/r/floibach/github-repo-raw-urls-ui)


**GitHub Raw URL Explorer** is a small web UI that scans a public GitHub repository and lists **raw.githubusercontent.com URLs** for all files in the selected branch.

---

This project is built and maintained in my free time.  
If it helps you or saves you some time, you can support my work on [![BuyMeACoffee](https://raw.githubusercontent.com/pachadotdev/buymeacoffee-badges/main/bmc-black.svg)](https://buymeacoffee.com/floibach)

Thank you for your support!

## What is this?

The main idea behind this tool is to make it easier to work with **AI chatbots / LLMs**.

When chatting with AI systems, you usually cannot paste an entire repository:
- too large
- too noisy
- easy to miss files

However, most AI tools *can* fetch and read **raw GitHub file URLs**.

This service helps you quickly generate a **complete list of raw file links**, so you can paste them into an AI chat and let the model:
- write or improve tests
- review or polish UI code
- generate documentation
- refactor or explain code
- reason about the whole repository structure

In short:  
**Give the AI the repo via links instead of copy & paste.**

## Screenshots

### Repository scan
![Overview](https://raw.githubusercontent.com/florianibach/github-repo-raw-urls-ui/refs/heads/master/docs/screenshots/overview.png)


## Features

- Scan any **public GitHub repository**
- Select branch (default branch is detected automatically)
- Recursively lists all files
- Generates `raw.githubusercontent.com` URLs
- Copy all links with one click
- No authentication required
- No GitHub token needed
- Lightweight and self-contained

---

## Usage

1. Enter a public GitHub repository URL  
   Example:
```

https://github.com/florianibach/github-repo-raw-urls-ui

````

2. Optionally select a branch  
(the default branch is marked)

3. Click **Scan**

4. Copy the generated raw URLs and paste them into your AI chat

---

## Docker

### docker run

```bash
docker run -d \
--name github-repo-raw-urls-ui \
-p 8080:8080 \
--restart unless-stopped \
floibach/github-repo-raw-urls-ui
````

Open in your browser:

```
http://localhost:8080
```

---

### docker compose

```yaml
version: "3.8"

services:
  github-repo-raw-urls-ui:
    image: floibach/github-repo-raw-urls-ui
    container_name: github-repo-raw-urls-ui
    ports:
      - "8080:8080"
    restart: unless-stopped
```

Start the service:

```bash
docker compose up -d
```

---

## Technical notes

* Written in Go
* Uses the GitHub REST API (unauthenticated)
* No data is stored
* No analytics
* No tracking
* All processing happens at request time

---

## Limitations

* Only public repositories are supported
* GitHub API rate limits apply (unauthenticated requests)
* Very large repositories may take longer to scan

---

## Disclaimer

This tool was developed in close collaboration with an AI chat assistant and refined iteratively through human–AI interaction.

The final design decisions, implementation, and maintenance remain entirely human-driven.

---

## License

MIT License
