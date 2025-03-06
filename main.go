package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	WebhookURL string `json:"webhook_url"`
	Repos      []Repo `json:"repos"`
}

type Repo struct {
	RepoURL   string   `json:"repo_url"`
	RepoPath  string   `json:"repo_path"`
	Branch    string   `json:"branch"`
	Commands  []string `json:"commands"`
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

			err := stashAndFetch(repo.RepoPath, repo.Branch)
			if err != nil {
				logError(config.WebhookURL, fmt.Sprintf("❌ Erreur lors du stash et fetch pour %s: %v", repo.RepoURL, err))
				continue
			}

			if hasUpdates(repo.RepoPath, repo.Branch) {
				fmt.Println("🚀 Mise à jour détectée, pull en cours...")
				_, err := runCommand(repo.RepoPath, "git", "pull", "origin", repo.Branch)
				if err != nil {
					logError(config.WebhookURL, fmt.Sprintf("❌ Erreur lors du pull pour %s: %v", repo.RepoURL, err))
					continue
				}

				sendDiscordEmbedWebhook(config.WebhookURL, repo.RepoURL, repo.RepoPath, repo.Branch, repo.Commands)
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
	cmdArgs := append([]string{"/C", "cd", dir, "&&", name}, args...)

	cmd := exec.Command("cmd", cmdArgs...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

func stashAndFetch(repoPath string, branch string) error {
	_, err := runCommand(repoPath, "git", "stash")
	if err != nil {
		return fmt.Errorf("erreur lors du stash des modifications : %v", err)
	}

	_, err = runCommand(repoPath, "git", "fetch", "origin")
	if err != nil {
		return fmt.Errorf("erreur lors du fetch : %v", err)
	}

	return nil
}

func runCustomCommands(repoPath string, commands []string) (string, error) {
	var output bytes.Buffer

	for _, cmd := range commands {
		args := strings.Split(cmd, " ")
		cmdOutput, err := runCommand(repoPath, args[0], args[1:]...)
		if err != nil {
			return output.String(), fmt.Errorf("❌ Erreur lors de l'exécution de la commande `%s`: %v\nOutput: %s", cmd, err, cmdOutput)
		}
		output.WriteString(fmt.Sprintf("✅ Commande `%s` exécutée avec succès.\nOutput:\n%s\n", cmd, cmdOutput))
	}

	return output.String(), nil
}

func sendDiscordEmbedWebhook(webhookURL, repoURL, repoPath, branch string, commands []string) {
	if webhookURL == "" {
		return
	}

	oldCommit, _ := runCommand(repoPath, "git", "rev-parse", "HEAD")
	oldCommit = strings.TrimSpace(oldCommit)

	pullOutput, _ := runCommand(repoPath, "git", "pull", "origin", branch)

	newCommit, _ := runCommand(repoPath, "git", "rev-parse", "HEAD")
	newCommit = strings.TrimSpace(newCommit)

	commandsOutput := ""
	if len(commands) > 0 {
		output, err := runCustomCommands(repoPath, commands)
		if err != nil {
			commandsOutput = fmt.Sprintf("❌ Erreur lors de l'exécution des commandes : %v", err)
		} else {
			commandsOutput = output
		}
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       "🚀 Mise à jour détectée et appliquée",
				"description": fmt.Sprintf("Le dépôt **[%s]** a été mis à jour sur la branche `%s`.", repoURL, branch),
				"color":       5814783,
				"fields": []map[string]string{
					{
						"name":  "Ancien Commit",
						"value": fmt.Sprintf("`%s`", oldCommit),
					},
					{
						"name":  "Nouveau Commit",
						"value": fmt.Sprintf("`%s`", newCommit),
					},
					{
						"name":  "Logs du Pull",
						"value": fmt.Sprintf("```%s```", truncateString(pullOutput, 1000)),
					},
					{
						"name":  "Logs des Commandes",
						"value": fmt.Sprintf("```%s```", truncateString(commandsOutput, 1000)),
					},
				},
				"timestamp": time.Now().Format(time.RFC3339),
			},
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("❌ Erreur lors de l'envoi du webhook Discord:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		fmt.Println("❌ Erreur: Webhook Discord a retourné un statut inattendu:", resp.Status)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func logError(webhookURL string, message string) {
	fmt.Println(message)
	sendDiscordEmbedWebhook(webhookURL, "Erreur", "N/A", "N/A", []string{})
}
