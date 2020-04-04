# 🕷 makescraper

[![Go Report Card](https://goreportcard.com/badge/github.com/omarsagoo/makescraper_v1)](https://goreportcard.com/report/github.com/omarsagoo/makescraper_v1)


### 📚 Table of Contents

1.[Description](#description)
2. [Project Structure](#project-structure)
3. [Getting Started](#getting-started)

## Description

This web scraper grabs all of the articles and related articles off of the google news web page, stores it into JSON and adds the articles (without duplicates) to a database. 

## Project Structure

```bash
📂 makescraper
├── README.md
└── scrape.go
```

## Getting Started

1. Create an empty repository to store this file.
2. Run each command line-by-line in your terminal to set up the project:

    ```bash
    $ git clone git@github.com:omarsagoo/makescraper_v1.git
    $ cd makescraper_v1
    $ git remote rm origin
    $ git remote add origin git@github.com:YOUR_GITHUB_USERNAME/YOUR_REPO_NAME.git
    $ go mod download
    ```
3. Run this command to start scraping:

    ```bash
    $ go build && ./makescraper_v1
    ```

