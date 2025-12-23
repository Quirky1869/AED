# AED

![](./_images/aed.png)  

![Static Badge](https://img.shields.io/badge/TUI-AED-gold?style=plastic)
![Static Badge](https://img.shields.io/badge/License-MIT-500aff?style=plastic)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=plastic&logo=go)](https://golang.org/)
[![Latest Version](https://img.shields.io/github/v/release/Quirky1869/aed?color=b100ff)](https://github.com/Quirky1869/aed/releases)
[![GitHub Actions Release](https://github.com/Quirky1869/aed/actions/workflows/release.yml/badge.svg)](https://github.com/Quirky1869/aed/actions/workflows/release.yml)
[![GitHub Actions Translate](https://github.com/Quirky1869/aed/actions/workflows/translate-en-readme.yml/badge.svg)](https://github.com/Quirky1869/aed/actions/workflows/translate-en-readme.yml)

## English README :gb:

An English version of the README is available [here](README-en.md)  

## Résumé
<p align="center">  
<a href="https://golang.org" target="_blank" rel="noreferrer">  
  <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/go/go-original.svg" alt="go" width="40" height="70"/>  
</a>  
</p>   

AED (Analyseur d'Espace Disque) est un explorateur de disque interactif pour terminal. Il scanne récursivement un dossier donné pour afficher quels fichiers ou sous-dossiers consomment le plus d'espace.  

## Fonctionnalités

Analyse intelligente : Scan récursif qui trie les éléments par taille, nom ou nombre d'éléments et gère les hardlinks (via inodes).  

Visualisation claire : Affiche des barres de progression colorées pour identifier instantanément les éléments volumineux.  

Navigation ergonomique : Compatible avec les flèches directionnelles (↑/↓/←/→) et les touches Vim (h, j, k, l) pour naviguer dans l'arborescence.  

>[!TIP]  
> Options :
> - ? : Afficher l'aide
> - esc : Revenir au menu principal
> - b : Ouvre un shell (terminal) directement dans le dossier sélectionné  
> - g : Ouvre le dossier ou fichier dans l'explorateur graphique de l'OS (xdg-open)  
> - r : Permet de re-scanner le dossier actuel
> - ctrl+l : Permet de permuter la langue entre Francais :fr: et Anglais :gb:
> - s : Trier par taille (Ascendant/Descendant)
> - n : Trier par nom (Ascendant/Descendant)
> - C : Trier par nombre d'éléments (Ascendant/Descendant)
> - e : Afficher ou masquer les fichiers cachés

## Structure du projet

```
.
├── go.mod
├── cmd/
│   └── app/
│       └── main.go      # Entrée du programme
├── scanner/
│   └── scanner.go       # Logique de scan, FileNode, utilitaires
└── ui/
    ├── autocomplete.go  # Fichier d'autocompletion (Tab)
    ├── lang.go          # Fichier de langues (FR, EN)
    ├── model.go         # Structure Model, Update() et Init()
    ├── styles.go        # Styles Lipgloss
    └── view.go          # Fonction View() et formatage
```

## Installation

### Go.dev

Vous pouvez vous servir d'AED en le téléchargeant directement depuis [pkg.go.dev](https://pkg.go.dev/github.com/quirky1869/aed) :

`go install github.com/quirky1869/aed/cmd/aed@latest`  

Il vous faudra avoir Go d'installer sur votre PC  

<details>  
<summary style="font-weight: bold; color: #500aff;">Procédure d'installation Go (Cliquer pour déplier)</summary>  

```bash
# Supprimer une éventuelle ancienne version de Go
sudo rm -rf /usr/local/go  

# Dézipper le fichier téléchargé dans /usr/local
# (Adaptez le chemin et le nom du fichier si nécessaire)
sudo tar -C /usr/local -xzf ~/Téléchargements/go1.25.3.linux-amd64.tar.gz  

# Ajouter Go au PATH et définir GOPATH
# Ouvrir votre fichier ~/.bashrc ou ~/.zshrc
micro ~/.zshrc  

# Ajouter à la fin du fichier
export PATH=$PATH:/usr/local/go/bin  
export GOPATH=$HOME/go  
export PATH=$PATH:$GOPATH/bin  

# Recharger la configuration du shell
source ~/.zshrc  

# Vérifier la bonne installation
go version  
```
</details>  

>  Désinstallation  
>  `rm $(go env GOPATH)/bin/aed`

### Releases  

Les [releases](https://github.com/Quirky1869/aed/releases) sont disponibles [ici](https://github.com/Quirky1869/aed/releases)  

![](./_images/gif/aed.gif)  

### Snap

Via snapstore  

En cours ...  🛠️  

## Couleurs

Vous pouvez changer les couleurs de l'interface à partir du fichier `ui/styles.go`  

<details>  
<summary style="font-weight: bold; color: #500aff;">ui/styles.go (Cliquer pour déplier)</summary>  

### Selection du dossier et exclusions

<img src="./_images/DSA1.png" alt="DSA1">  
  
> - 1 = titleStyle
> - 2 = helpDescStyle
> - 3 = inactiveStyle
> - 4 = helpDescStyle

### Scan
<img src="./_images/DSA2.png" alt="DSA2">  

> - 1 = spinnerStyle
> - 2 = countStyle
> - 3 = helpDescStyle

### AED

<img src="./_images/DSA3.png" alt="DSA3">  

> - 1 = sortStyle
> - 2 = infoStyle
> - 3 = infoStyle
> - 4 = pathStyle
> - 5 = selectedStyle
> - 6 = barFull
> - 7 = barEmpty
> - 8 = helpKeyStyle
> - 9 = helpDescStyle
> - 10 = hiddenStyle
> - 11 = fileCountStyle

</details>  

## Technologies utilisées

| Librairie                                                    | Utilisation                          |
| ------------------------------------------------------------ | ------------------------------------ |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea)     | Gestion de l’interface TUI           |
| [Lipgloss](https://github.com/charmbracelet/lipgloss)        | Stylisation du texte et des bordures |
| ![Go](https://img.shields.io/badge/Go-%2300ADD8.svg?style=flat&logo=go&logoColor=white) [Golang](https://go.dev)   | Langage principal et logique métier |


## Auteur

Projet développé par Quirky  

<a href="https://github.com/Quirky1869" target="_blank">  
  <img src="./_images/white-github.png" alt="GitHub" width="30" height="30" style="vertical-align:middle;"> GitHub  
</a>  
