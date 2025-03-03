package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	RepoURL  string `json:"repo_url"`
	RepoPath string `json:"repo_path"`
	GitToken string `json:"git_token"`
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Println("❌ Erreur: Impossible de charger config.json:", err)
		os.Exit(1)
	}

	if config.GitToken == "" {
		fmt.Println("❌ ERREUR: Le token GitHub est manquant dans config.json.")
		os.Exit(1)
	}

	authURL := fmt.Sprintf("https://%s@%s", config.GitToken, config.RepoURL)

	for {
		fmt.Println("🔄 Vérification des mises à jour...")

		if hasUpdates(config.RepoPath, authURL) {
			fmt.Println("🚀 Mise à jour détectée, pull en cours...")
			runCommand(config.RepoPath, "git", "pull", authURL)

			runCommand(config.RepoPath, "pnpm", "install")
			runCommand(config.RepoPath, "pnpm", "run", "build")
		} else {
			fmt.Println("✅ Pas de mise à jour.")
		}

		time.Sleep(10 * time.Second)
	}
}

func loadConfig(filename string) (Config, error) {
	var config Config
	file, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(file, &config)
	return config, err
}

func hasUpdates(repoPath, authURL string) bool {
	output := runCommand(repoPath, "git", "fetch", authURL, "--dry-run")
	return strings.TrimSpace(output) != ""
}

func runCommand(dir string, name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Erreur lors de l'exécution de %s %v: %v\n", name, args, err)
	}

	return string(output)
}
