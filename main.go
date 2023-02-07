package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

type GitCommitData struct {
	hash    string
	date    string
	author  string
	message string
}

func main() {
	app := tview.NewApplication()
	list := tview.NewList()

	commits := getGitLog()
	for _, commit := range commits {
		list.AddItem(fmt.Sprintf("- %s - %s - %s", commit.hash, commit.author, commit.date), commit.message, 'a', nil)
	}
	list.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})
	list.SetBorder(true)

	if err := app.SetRoot(list, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}
}

var DIR = "/home/walker/personal/temp1"

func getGitLog() []GitCommitData {
	dir := exec.Command("pwd")
	dir.Dir = DIR
	dir.Output()
	fmt.Println(dir.Stdout)

	git, _ := exec.LookPath("git")
	formatString := "%h|%an|%ar|%s"
	logCommand := exec.Command(git, "log", fmt.Sprintf("--pretty=format:%s", formatString), "--date=short")
	logCommand.Dir = DIR
	outputByte, err := logCommand.Output()
	if err != nil {
		fmt.Println(string(outputByte))
		panic(err)
	}

	commits := make([]GitCommitData, 0)

	reader := bytes.NewReader(outputByte)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		// fmt.Println(line)

		splits := strings.Split(line, "|")
		commit := GitCommitData{
			hash:    splits[0],
			author:  splits[1],
			date:    splits[2],
			message: splits[3],
		}
		commits = append(commits, commit)
	}
	fmt.Println(commits)
	return commits
}
