#AED

![](./_images/aed.jpeg)

![Static Badge](https://img.shields.io/badge/TUI-AED-gold?style=plastic)
![Static Badge](https://img.shields.io/badge/License-MIT-500aff?style=plastic)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=plastic&logo=go)](https://golang.org/)
[![Latest Version](https://img.shields.io/github/v/release/Quirky1869/aed?color=b100ff)](https://github.com/Quirky1869/aed/releases)
[![GitHub Actions Release](https://github.com/Quirky1869/aed/actions/workflows/build-and-release.yml/badge.svg)](https://github.com/Quirky1869/aed/actions/workflows/build-and-release.yml)

## English README :gb:

An English version of the README is available [here](README-en.md)

## Summary
<p align="center">
<a href="https://golang.org" target="_blank" rel="noreferrer">
<img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/go/go-original.svg" alt="go" width="40" height="70"/>
</a>
</p>

AED (Disk Space Analyzer) is an interactive disk explorer for terminal. It recursively scans a given folder to show which files or subfolders are consuming the most space.

## Features

Smart scan: Recursive scan that sorts elements by size, name or number of elements and manages hardlinks (via inodes).

Clear visualization: Shows colored progress bars to instantly identify large items.

Ergonomic navigation: Compatible with directional arrows (↑/↓/←/→) and Vim keys (h, j, k, l) to navigate the tree.

>[!TIP]
> Options:
> - ? : Show help
> - esc: Return to main menu
> - b: Opens a shell (terminal) directly in the selected folder.
> - g: Opens the folder or file in the OS graphical explorer (xdg-open).
> - r: Allows you to re-scan the current folder
> - ctrl+l: Allows you to switch the language between French:fr: and English:gb:
> - s: Sort by size (Ascending/Descending)
> - n: Sort by name (Ascending/Descending)
> - C: Sort by number of elements
> - h: Show or hide hidden files

## Project structure

```
.
├── go.mod
├── cmd/
│ └── app/
│   └── main.go # Program entry
├── scan/
│ └── scanner.go # Scan logic, FileNode, utilities
└── ui/
├── lang.go # Language file (FR, EN)
├── styles.go # Lipgloss Styles
├── view.go # View() function and formatting
└── model.go # Structure Model, Update() and Init()
```

## Releases

The [releases](https://github.com/Quirky1869/aed/releases) are available [here](https://github.com/Quirky1869/aed/releases)

![](./_images/gif/aed.gif)

## Technologies used

| Bookstore | Use |
| ------------------------------------------------------------ | ------------------------------------ |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Management of the TUI interface |
| [Lipgloss](https://github.com/charmbracelet/lipgloss) | Stylizing text and borders |
| ![Go](https://img.shields.io/badge/Go-%2300ADD8.svg?style=flat&logo=go&logoColor=white) [Golang](https://go.dev) | Core language and business logic |


## Author

Project developed by Quirky

<a href="https://github.com/Quirky1869" target="_blank">
<img src="./_images/white-github.png" alt="GitHub" width="30" height="30" style="vertical-align:middle;"> GitHub
</a>
