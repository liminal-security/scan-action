package git

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"slices"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Differ struct {
	repo        *git.Repository
	shallowEnds []string
}

func NewDiffer(repoPath string) (differ *Differ, err error) {
	gitRepo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("can't open repo %s: %s", repoPath, err)
	}

	shallowEnds, err := readShallow(repoPath)
	if err != nil {
		return nil, fmt.Errorf("can't read git/shallow: %w", err)
	}

	return &Differ{
		repo:        gitRepo,
		shallowEnds: shallowEnds,
	}, nil
}

func readShallow(repoPath string) (commitsSHA []string, err error) {
	file, err := os.OpenFile(path.Join(repoPath, ".git/shallow"), os.O_RDONLY, 0)
	if errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	}
	defer file.Close()

	commitsSHA = []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		commitsSHA = append(commitsSHA, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return []string{}, fmt.Errorf("scan error: %w", err)
	}

	return commitsSHA, scanner.Err()
}

func (d *Differ) Diff() (commits []Commit, err error) {
	cIter, err := d.repo.Log(&git.LogOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("error creating commit iterator: %w", err)
	}

	err = cIter.ForEach(func(c *object.Commit) error {
		commit := Commit{
			Hash: c.Hash.String(),
		}
		diff := Diff{}
		diff.Data = map[string]string{}

		if slices.Contains(d.shallowEnds, c.Hash.String()) {
			return nil
		}

		commitTree, err := c.Tree()
		if err != nil {
			return fmt.Errorf("error getting commit tree: %w", err)
		}

		parentTree, err := getParent(c)
		if err != nil {
			return fmt.Errorf("can't get commit parent %w", err)
		}

		patch, err := parentTree.Patch(commitTree)
		if err != nil {
			return fmt.Errorf("error getting patch: %w", err)
		}

		filePatches := patch.FilePatches()
		for _, p := range filePatches {
			filePath, err := getPath(p.Files())
			if err != nil {
				return fmt.Errorf("error getting file path: %w", err)
			}

			var data string
			chunks := p.Chunks()
			for _, c := range chunks {
				data += c.Content()
			}

			diff.Data[filePath] = data
			commit.Diff = diff
			commits = append(commits, commit)
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return commits, nil
}

func getParent(c *object.Commit) (tree *object.Tree, err error) {
	if c.NumParents() != 0 {
		parent, err := c.Parents().Next()
		if err != nil {
			return nil, fmt.Errorf("can't get commit parent: %w", err)
		}

		parentTree, err := parent.Tree()
		if err != nil {
			return nil, fmt.Errorf("can't get commit parent tree: %w", err)
		}

		return parentTree, nil
	} else {
		return &object.Tree{}, nil
	}
}

func getPath(from, to diff.File) (string, error) {
	if isNilFile(from) && isNilFile(to) {
		return "", fmt.Errorf("can't determine path")
	}

	if !isNilFile(to) {
		return to.Path(), nil
	}

	return from.Path(), nil
}

func isNilFile(f diff.File) bool {
	if f == nil {
		return true
	}
	v := reflect.ValueOf(f)

	return v.Kind() == reflect.Ptr && v.IsNil()
}
