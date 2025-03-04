package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	"bytes"
)

type Config struct {
	WebhookURL string `json:"webhook_url"`
	Repos      []Repo `json:"repos"`
}

type Repo struct {
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

		for _, repo := range config.Repos {
			fmt.Printf("🔍 Vérification du dépôt : %s (%s)\n", repo.RepoURL, repo.Branch)

            _, err := runCommand(repo.RepoPath, "git", "fetch", "origin")
            if err != nil {
                logError(config.WebhookURL, fmt.Sprintf("❌ Erreur lors du fetch pour %s: %v", repo.RepoURL, err))
                continue
            }


			if hasUpdates(repo.RepoPath, repo.Branch) {
				fmt.Println("🚀 Mise à jour détectée, pull en cours...")
				_, err := runCommand(repo.RepoPath, "git", "pull", "origin", repo.Branch)
				if err != nil {
					logError(config.WebhookURL, fmt.Sprintf("❌ Erreur lors du pull pour %s: %v", repo.RepoURL, err))
					continue
				}

				fmt.Println("🔨 Installation des dépendances et build en cours...")
				_, err = runCommand(repo.BuildPath, "pnpm", "install")

				if err != nil {
					logError(config.WebhookURL, fmt.Sprintf("❌ Erreur lors de l'installation des dépendances pour %s: %v", repo.RepoURL, err))
					continue
				}

				_, err = runCommand(repo.BuildPath, "pnpm", "run", "build")

				if err != nil {
					logError(config.WebhookURL, fmt.Sprintf("❌ Erreur lors du build pour %s: %v", repo.RepoURL, err))
					continue
				}
				sendDiscordWebhook(config.WebhookURL, fmt.Sprintf("🚀 Mise à jour détectée et build réussi pour le dépôt : %s (%s)" + "", repo.RepoURL, repo.Branch))
			} else {
				fmt.Println("✅ Pas de mise à jour.")
			}
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
	localCommit, err := runCommand(repoPath, "git", "rev-parse", "HEAD")
	if err != nil {
		return false
	}
	localCommit = strings.TrimSpace(localCommit)

	remoteCommit, err := runCommand(repoPath, "git", "rev-parse", fmt.Sprintf("origin/%s", branch))
	if err != nil {
		return false
	}
	remoteCommit = strings.TrimSpace(remoteCommit)

	return localCommit != remoteCommit
}

func runCommand(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

func sendDiscordWebhook(webhookURL string, message string) {
	if webhookURL == "" {
		return
	}

	payload := map[string]string{
		"content": message,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("❌ Erreur lors de l'envoi du webhook Discord:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		fmt.Println("❌ Erreur: Le webhook Discord a retourné un statut inattendu:", resp.Status)
	}
}

func logError(webhookURL string, message string) {
	fmt.Println(message)
	sendDiscordWebhook(webhookURL, message)
}