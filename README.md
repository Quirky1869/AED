# AED

![](./_images/aed.jpeg)  

![Static Badge](https://img.shields.io/badge/TUI-AED-gold?style=plastic)
![Static Badge](https://img.shields.io/badge/License-MIT-500aff?style=plastic)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=plastic&logo=go)](https://golang.org/)
[![Latest Version](https://img.shields.io/github/v/release/Quirky1869/aed?color=b100ff)](https://github.com/Quirky1869/aed/releases)
[![GitHub Actions Release](https://github.com/Quirky1869/aed/actions/workflows/build-and-release.yml/badge.svg)](https://github.com/Quirky1869/aed/actions/workflows/build-and-release.yml)

## English README :gb:

An English version of the readme is available [here](README-en.md)

## Résumé
<p align="center">  
<a href="https://golang.org" target="_blank" rel="noreferrer">  
  <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/go/go-original.svg" alt="go" width="40" height="70"/>  
</a>  
</p>   

AED (Analyseur d'Espace Disque) est un explorateur de disque interactif pour terminal. Il scanne récursivement un dossier donné pour afficher quels fichiers ou sous-dossiers consomment le plus d'espace.  

## Fonctionnalités

Analyse intelligente : Scan récursif qui trie les éléments par taille et gère les hardlinks (via inodes).  

Visualisation claire : Affiche des barres de progression colorées pour identifier instantanément les éléments volumineux.  

Navigation ergonomique : Compatible avec les flèches directionnelles et les touches Vim (h, j, k, l) pour naviguer dans l'arborescence.  

>[!TIP]
> Options :
> - b : Ouvre un shell (terminal) directement dans le dossier sélectionné.  
> - g : Ouvre le dossier ou fichier dans l'explorateur graphique de l'OS (xdg-open).  
> - r : Permet de re-scanner le dossier actuel
> - ctrl+l : Permet de permuter la langue entre Francais :fr: et Anglais :gb:
> - s : Trier par taille (Ascendant/Descendant)
> - n : Trier par nom (Ascendant/Descendant)
> - C : Trier par nombre d'éléments
> - h : Afficher ou masquer les fichiers cachés

## Structure du projet

```
.
├── go.mod
├── cmd/
│   └── app/
│       └── main.go  # Entrée du programme
├── scanner/
│   └── scanner.go   # Logique de scan, FileNode, utilitaires
└── ui/
    ├── lang.go      # Fichier de langues (FR, EN)
    ├── styles.go    # Styles Lipgloss
    ├── view.go      # Fonction View() et formatage
    └── model.go     # Structure Model, Update() et Init()
```

## Releases

Les [releases](https://github.com/Quirky1869/aed/releases) sont disponibles [ici](https://github.com/Quirky1869/aed/releases)  

![](./_images/gif/aed.gif)  

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
