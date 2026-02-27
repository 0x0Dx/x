package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	Success = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	Error   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	Info    = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	Dim     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)
