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
	RepoURL   string `json:"repo_url"`
	RepoPath  string `json:"repo_path"`
	BuildPath string `json:"build_path"`
	Branch    string `json:"branch"`
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Println("❌ Erreur: Impossible de charger config.json:", err)
		os.Exit(1)
	}

	for {
		fmt.Println("🔄 Vérification des mises à jour...")

		fmt.Println("🔍 Fetch des dernières modifications...")
		runCommand(config.RepoPath, "git", "fetch", "origin")

		if hasUpdates(config.RepoPath, config.Branch) {
			fmt.Println("🚀 Mise à jour détectée, pull en cours...")
			runCommand(config.RepoPath, "git", "pull", "origin", config.Branch)

			fmt.Println("🔨 Installation des dépendances et build en cours...")
			runCommand(config.BuildPath, "pnpm", "install")
			runCommand(config.BuildPath, "pnpm", "run", "build")
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

func hasUpdates(repoPath, branch string) bool {
	localCommit := runCommand(repoPath, "git", "rev-parse", "HEAD")
	localCommit = strings.TrimSpace(localCommit)

	remoteCommit := runCommand(repoPath, "git", "rev-parse", fmt.Sprintf("origin/%s", branch))
	remoteCommit = strings.TrimSpace(remoteCommit)

	return localCommit != remoteCommit
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