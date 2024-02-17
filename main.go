package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
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

	APP.SetRoot(list, true).SetFocus(list)
}

func updateCommitTime(gcd GitCommitData, list *tview.List) {
	form := tview.NewForm().
		AddTextView("Hash", gcd.hash, 40, 1, true, false).
		AddTextView("Message", gcd.message, 40, 1, true, false).
		AddInputField("Date", gcd.date, 30, nil, nil)

	form.AddButton("Save", func() {
		dateField := form.GetFormItemByLabel("Date").(*tview.InputField)
		newDate := dateField.GetText()
		newDate = parseToGitTimestamp(newDate)

		commandText := fmt.Sprintf(
			`
if commit.original_id.startswith(b"%s"):
  commit.author_date = b"%s";
  commit.committer_date = b"%s";
`,
			gcd.hash, newDate, newDate)

		command := exec.Command("git", "filter-repo", "--force", "--commit-callback", commandText)
		output, err := command.Output()

		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				log.Println("err is", string(ee.Stderr))
			}
		} else {
			log.Println("completed", string(output))
		}
		populateCommitList(list)
	})

	form.
		SetTitle("Update commit time").
		SetTitleAlign(tview.AlignLeft)

	APP.SetRoot(form, true).EnableMouse(true).Run()
}

func getGitLog() []GitCommitData {
	formatString := "%h|%an|%aI|%s"

	logCommand := exec.Command(git, "log", fmt.Sprintf("--pretty=format:%s", formatString), "--date=short")

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

func parseToGitTimestamp(dte string) string {
	t, err := time.Parse(time.RFC3339, dte)
	if err != nil {
		panic(err)
	}

	unix := t.Unix()
	_, offsetSeconds := t.Zone()
	offsetStr := zoneOffsetToString(offsetSeconds)

	data := fmt.Sprintf("%d %s", unix, offsetStr)
	log.Printf("input %s, unix %d, offset seconds %d is parsed into %s\n", dte, unix, offsetSeconds, data)
	return data
}

func zoneOffsetToString(offsetSeconds int) string {
	var result string

	hours := offsetSeconds / (60 * 60)
	minutes := offsetSeconds/60 - hours*60

	if offsetSeconds >= 0 {
		result = fmt.Sprintf("+%02d%02d", hours, minutes)
	} else {
		result = fmt.Sprintf("-%02d%02d", hours, minutes)
	}
	return result
}

func main() {
	fileName := time.Now().Format("20060102150405.log")

	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	list := tview.NewList()
	populateCommitList(list)

	if err := APP.SetRoot(list, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
