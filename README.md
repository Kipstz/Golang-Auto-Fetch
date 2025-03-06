## Description du projet

Ce projet est un **automate de mise à jour de dépôts Git** avec notifications sur Discord. Il permet de **vérifier périodiquement les mises à jour** d’un ou plusieurs dépôts Git, d’exécuter un `git pull` en cas de modification, puis d'envoyer un message sur un webhook Discord pour signaler l’opération. 

## Fonctionnalités principales

- **Chargement de la configuration** depuis un fichier JSON (`config.json`).
- **Vérification des mises à jour** sur plusieurs dépôts définis.
- **Stash des modifications locales** avant de récupérer les nouvelles modifications distantes.
- **Exécution d’un `git pull`** si des mises à jour sont détectées.
- **Exécution de commandes personnalisées** après la mise à jour (si définies dans la config).
- **Envoi d’une notification via un webhook Discord** avec les détails de la mise à jour.
- **Journalisation des erreurs** en cas de problème.

## Utilisation

1. **Configurer `config.json`** avec les URLs des dépôts, leurs chemins locaux et les éventuelles commandes à exécuter après la mise à jour.
2. **Lancer le programme** (go run main.go en mode développement ou exécuter le binaire compilé depuis la release`).
3. Le script s'exécute en boucle toutes les **10 secondes** pour vérifier les mises à jour.

## Cas d'usage

- **Déploiement automatique** sur un serveur après chaque mise à jour du dépôt.
- **Surveillance des mises à jour** pour des projets critiques.
- **Exécution de commandes spécifiques** (ex: redémarrage de services, compilation...).

C'est un outil pratique pour **automatiser la gestion de projets Git** et recevoir des notifications instantanées sur Discord. 🚀
