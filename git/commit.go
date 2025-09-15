package git

import (
	"fmt"
	"sort"
	"strings"
)

type Diff struct {
	Data map[string]string
}

type Commit struct {
	Hash string
	Diff Diff
}

func (c Commit) String() string {
	var b strings.Builder
	fileNames := make([]string, 0, len(c.Diff.Data))

	for fileName := range c.Diff.Data {
		fileNames = append(fileNames, fileName)
	}

	sort.Strings(fileNames)

	for _, fileName := range fileNames {
		b.WriteString(c.Diff.Data[fileName])
		b.WriteString("\n")
	}

	return b.String()
}

func (c Commit) GetFileNameByLine(lineNum int) (fileName string, err error) {
	if lineNum < 1 {
		return "", fmt.Errorf("invalid line number %d", lineNum)
	}

	fileNames := make([]string, 0, len(c.Diff.Data))

	for fileName := range c.Diff.Data {
		fileNames = append(fileNames, fileName)
	}

	sort.Strings(fileNames)

	totalNewLines := 0

	for _, fileName := range fileNames {
		fileData := c.Diff.Data[fileName] + "\n"
		newlineCount := strings.Count(fileData, "\n")

		totalNewLines += newlineCount

		if totalNewLines+1 > lineNum {
			return fileName, nil
		}
	}

	return "", fmt.Errorf("no data found for line number %d", lineNum)
}
