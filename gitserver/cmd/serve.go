package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0x0Dx/x/gitserver/models"
)

var (
	COMMANDS_READONLY = map[string]int{
		"git-upload-pack": models.AU_WRITABLE,
		"git upload-pack": models.AU_WRITABLE,
	}

	COMMANDS_WRITE = map[string]int{
		"git-receive-pack": models.AU_READABLE,
		"git receive-pack": models.AU_READABLE,
	}
)

func In(b string, sl map[string]int) bool {
	_, e := sl[b]
	return e
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run git server",
	Args:  cobra.ArbitraryArgs,
	Run: func(cc *cobra.Command, args []string) {
		keys := strings.Split(os.Args[2], "-")
		if len(keys) != 2 {
			fmt.Println("auth file format error")
			return
		}

		keyId, err := strconv.ParseInt(keys[1], 10, 64)
		if err != nil {
			fmt.Println("auth file format error")
			return
		}
		user, err := models.GetUserByKeyId(keyId)
		if err != nil {
			fmt.Println("You have no rights to access")
			return
		}

		cmd := os.Getenv("SSH_ORIGINAL_COMMAND")
		if cmd == "" {
			fmt.Printf("Hi %s! You've successfully authenticated, but gitserver does not support shell access.\n", user.Name)
			return
		}

		f, _ := os.Create("test2.log")
		f.WriteString(cmd)
		f.Close()

		log.Printf("cmd is %s", cmd)

		verb, repoArgs := parseCmd(cmd)
		rr := strings.SplitN(strings.Trim(repoArgs, "'"), "/", 1)
		if len(rr) != 2 {
			fmt.Println("Unavailable repository")
			return
		}
		repoName := rr[1]
		if strings.HasSuffix(repoName, ".git") {
			repoName = repoName[:len(repoName)-4]
		}
		isWrite := In(verb, COMMANDS_WRITE)
		isRead := In(verb, COMMANDS_READONLY)
		switch {
		case isWrite:
			has, err := models.HasAccess(user.Name, repoName, COMMANDS_WRITE[verb])
			if err != nil {
				fmt.Println("Internal error")
				return
			}
			if !has {
				fmt.Println("You have no rights to access this repository")
				return
			}
		case isRead:
			has, err := models.HasAccess(user.Name, repoName, COMMANDS_READONLY[verb])
			if err != nil {
				fmt.Println("Internal error")
				return
			}
			if !has {
				has, err = models.HasAccess(user.Name, repoName, COMMANDS_WRITE[verb])
				if err != nil {
					fmt.Println("Internal error")
					return
				}
			}
			if !has {
				fmt.Println("You have no right to access this repository")
				return
			}
		default:
			fmt.Println("Unknown command")
			return
		}

		isExist, err := models.IsRepositoryExist(user, repoName)
		if err != nil {
			fmt.Println("Internal error")
			return
		}

		if !isExist {
			if isRead {
				fmt.Println("Repository does not exist")
				return
			} else if isWrite {
				_, err := models.CreateRepository(user, repoName)
				if err != nil {
					fmt.Println("Create repository failed")
					return
				}
			}
		}

		fullPath := filepath.Join(models.RepoRootPath, user.Name, repoName+".git")
		newcmd := fmt.Sprintf("%s '%s'", verb, fullPath)
		fmt.Println(newcmd)
		gitcmd := exec.Command("git", "shell", "-c", newcmd)
		gitcmd.Stdout = os.Stdout
		gitcmd.Stderr = os.Stderr

		err = gitcmd.Run()
		if err != nil {
			log.Printf("Execute command error: %s", err)
		}
	},
}

func parseCmd(cmd string) (string, string) {
	ss := strings.SplitN(cmd, " ", 1)
	if len(ss) != 2 {
		return "", ""
	}
	verb, args := ss[0], ss[1]
	if verb == "git" {
		ss = strings.SplitN(args, " ", 1)
		if len(ss) != 2 {
			return "", ""
		}
		args = ss[1]
		verb = fmt.Sprintf("%s %s", verb, ss[0])
	}
	return verb, args
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
