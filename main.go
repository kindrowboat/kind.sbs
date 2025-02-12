package main

import (
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Condition struct {
	Name               string   `yaml:"name"`
	Emoji              string   `yaml:"emoji"`
	Description        string   `yaml:"description"`
	ExpectedBehaviours []string `yaml:"expected_behaviours"`
	FriendshipSupport  []string `yaml:"friendship_support"`
}

type Manifest struct {
	Title       string      `yaml:"title"`
	Description string      `yaml:"description"`
	Conditions  []Condition `yaml:"conditions"`
}

var manifest Manifest

func main() {
	// Load and parse the manifest.yaml file
	data, err := os.ReadFile("manifest.yaml")
	if err != nil {
		log.Fatalf("Failed to read manifest.yaml: %v", err)
	}
	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		log.Fatalf("Failed to parse manifest.yaml: %v", err)
	}

	// Parse HTML templates
	tmpl := template.Must(template.ParseFiles(
		"templates/index.html.tmpl",
		"templates/condition.html.tmpl",
	))

	// Create the public directory if it doesn't exist
	err = os.MkdirAll("public", os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create public directory: %v", err)
	}

	// Copy the static assets from the assets directory
	err = filepath.Walk("assets", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			destPath := filepath.Join("public", path[len("assets/"):])
			err = os.MkdirAll(filepath.Dir(destPath), os.ModePerm)
			if err != nil {
				return err
			}
			_, err = copyFile(path, destPath)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to copy assets: %v", err)
	}

	// Generate the index page
	indexFile, err := os.Create(filepath.Join("public", "index.html"))
	if err != nil {
		log.Fatalf("Failed to create index.html: %v", err)
	}
	defer indexFile.Close()
	err = tmpl.ExecuteTemplate(indexFile, "index.html.tmpl", manifest)
	if err != nil {
		log.Fatalf("Failed to execute template for index.html: %v", err)
	}

	// Generate condition pages
	for _, condition := range manifest.Conditions {
		conditionDir := filepath.Join("public", condition.Emoji)
		err = os.MkdirAll(conditionDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create directory %s: %v", conditionDir, err)
		}
		conditionFile, err := os.Create(filepath.Join(conditionDir, "index.html"))
		if err != nil {
			log.Fatalf("Failed to create %s/index.html: %v", conditionDir, err)
		}
		defer conditionFile.Close()
		err = tmpl.ExecuteTemplate(conditionFile, "condition.html.tmpl", struct {
			Manifest  Manifest
			Condition Condition
		}{
			Manifest:  manifest,
			Condition: condition,
		})
		if err != nil {
			log.Fatalf("Failed to execute template for %s/index.html: %v", conditionDir, err)
		}
	}

	log.Println("Site generated successfully in the public directory")
}

func copyFile(src, dst string) (int64, error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destFile.Close()

	nBytes, err := io.Copy(destFile, sourceFile)
	return nBytes, err
}
