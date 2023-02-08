package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/rivo/tview"
)

type GitCommitData struct {
	hash    string
	date    string
	author  string
	message string
}

var (
	APP    = tview.NewApplication()
	git, _ = exec.LookPath("git")
)

func populateCommitList(list *tview.List) {
	list.Clear()

	commits := getGitLog()
	for _, commit := range commits {
		list.AddItem(fmt.Sprintf("- %s - %s - %s", commit.hash, commit.author, commit.date), commit.message, '0', func() {
			updateCommitTime(commit, list)
		})
	}
	list.AddItem("Quit", "Press to exit", 'q', func() {
		APP.Stop()
	})

	list.SetTitle("Commits").SetBorder(true)

	if err := APP.SetRoot(list, true).SetFocus(list).Run(); err != nil {
		panic(err)
	}
}

var DIR = "/home/walker/personal/temp1"

func updateCommitTime(gcd GitCommitData, list *tview.List) {
	form := tview.NewForm().
		AddInputField("Hash", gcd.hash, 20, nil, nil).
		AddInputField("Date", gcd.date, 20, nil, nil).
		AddInputField("Message", gcd.message, 20, nil, nil)

	form.AddButton("Save", func() {
		dateField := form.GetFormItemByLabel("Date").(*tview.InputField)
		newDate := dateField.GetText()

		commandText := fmt.Sprintf(
			`
if commit.original_id.startswith(b"%s"):
  commit.author_date = b"%s";
  commit.committer_date = b"%s";
`,
			gcd.hash, newDate, newDate)

		// fmt.Println("command", commandText)

		command := exec.Command(git, commandText)
		command.Output()

		// if err != nil {
		// 	println("fuck it failed", err.Error())
		// } else {
		// 	println("completed", command.Stdout)
		// }
		populateCommitList(list)
	})

	form.
		SetTitle("Update commit time").
		SetTitleAlign(tview.AlignLeft)

	if err := APP.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func getGitLog() []GitCommitData {
	dir := exec.Command("pwd")
	dir.Dir = DIR
	dir.Output()
	fmt.Println(dir.Stdout)

	formatString := "%h|%an|%ai|%s"

	logCommand := exec.Command(git, "log", fmt.Sprintf("--pretty=format:%s", formatString), "--date=short")
	logCommand.Dir = DIR

	outputByte, err := logCommand.Output()
	if err != nil {
		panic(err)
	}

	commits := make([]GitCommitData, 0)

	reader := bytes.NewReader(outputByte)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()

		splits := strings.Split(line, "|")
		commit := GitCommitData{
			hash:    splits[0],
			author:  splits[1],
			date:    splits[2],
			message: splits[3],
		}
		commits = append(commits, commit)
	}
	return commits
}

func parseToUnix(d string) string {
	t, err := time.Parse(time.RFC3339, d)
	if err != nil {
		panic(err)
	}

	unix := t.Unix()
	fmt.Println(unix)
	return strconv.Itoa(int(unix))
}

func main() {
	list := tview.NewList()
	// box := tview.NewBox().SetTitle("Git commit time UI update")
	populateCommitList(list)

	// 	newDate := "2021-01-01T09:09:09Z"
	// 	newDate = parseToUnix(newDate)
	// 	newDate = "1946684799 +0000"
	//
	// 	commandText := fmt.Sprintf(
	// 		`
	// if commit.original_id.startswith(b"%s"):
	//   commit.author_date = b"%s";
	//   commit.committer_date = b"%s";
	// `,
	// 		"bcfb922", newDate, newDate)
	//
	// 	fmt.Println("command", commandText)

	// command := exec.Command("git", "filter-repo", "--force", "--commit-callback", commandText)
	// command.Dir = DIR
	// println("full command:", command.String())
	// out, err := command.Output()
	//
	// if err != nil {
	// 	if ee, ok := err.(*exec.ExitError); ok {
	// 		fmt.Println("err is", string(ee.Stderr))
	// 	}
	// 	panic(err)
	// } else {
	// 	println("completed", string(out))
	// }
}
