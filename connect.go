package main

import "strconv"

type sshFinishedMsg struct{ err error }

func sshArgs(p Profile) []string {
	var args []string
	if p.Port > 0 && p.Port != 22 {
		args = append(args, "-p", strconv.Itoa(p.Port))
	}
	if p.KeyPath != "" {
		args = append(args, "-i", p.KeyPath)
	}
	return append(args, p.User+"@"+p.Host)
}
