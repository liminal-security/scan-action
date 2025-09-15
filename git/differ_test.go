package git

import (
	"os"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/stretchr/testify/assert"
)

func TestOneCommit(t *testing.T) {
	path := checkout(t, "testdata/scan-action-test", "e86f19f49a18854efdbc753d2cc7c266fdcf6b5f", "1", 2)
	defer os.RemoveAll(path)

	differ, err := NewDiffer(path)
	if err != nil {
		t.Fatalf("Can't create differ: %s", err)
	}

	expectedCommits := []Commit{
		{
			Hash: "e86f19f49a18854efdbc753d2cc7c266fdcf6b5f",
			Diff: Diff{
				Data: map[string]string{
					"README.md": "# scan-action-test# scan-action-test\n\n\nAdded new stuff",
				},
			},
		},
	}

	commits, err := differ.Diff()
	if err != nil {
		t.Fatalf("Can't diff: %s", err)
	}

	assert.Equal(t, expectedCommits, commits)

}

func TestMultipleCommits(t *testing.T) {
	path := checkout(t, "testdata/scan-action-test", "539533aab24270f6201fcdd5aa25f6c16662ee58", "3", 3)
	defer os.RemoveAll(path)

	differ, err := NewDiffer(path)
	if err != nil {
		t.Fatalf("Can't create differ: %s", err)
	}

	expectedCommits := []Commit{
		{
			Hash: "539533aab24270f6201fcdd5aa25f6c16662ee58",
			Diff: Diff{
				Data: map[string]string{
					"notes.md": "# Notes# Notes\n\n## One more note",
				},
			},
		},
		{
			Hash: "9006ae9c5d2b99c774da25f7b91bd7e8457b2275",
			Diff: Diff{
				Data: map[string]string{
					"notes.md": "# Notes",
				},
			},
		},
	}

	commits, err := differ.Diff()
	if err != nil {
		t.Fatalf("Can't diff: %s", err)
	}

	assert.Equal(t, expectedCommits, commits)
}

func TestGetPath(t *testing.T) {
	tests := []struct {
		name    string
		from    *File
		to      *File
		wantRes string
		wantErr bool
	}{
		{
			name:    "empty",
			from:    nil,
			to:      nil,
			wantRes: "",
			wantErr: true,
		},
		{
			name:    "added",
			from:    nil,
			to:      &File{path: "file.txt"},
			wantRes: "file.txt",
			wantErr: false,
		},
		{
			name:    "removed",
			from:    &File{path: "file.txt"},
			to:      nil,
			wantRes: "file.txt",
			wantErr: false,
		},
		{
			name:    "renamed",
			from:    &File{path: "file.txt"},
			to:      &File{path: "other_file.txt"},
			wantRes: "other_file.txt",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := getPath(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRes != tt.wantRes {
				t.Errorf("getPath() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

type File struct {
	path string
}

func (f *File) Path() string {
	return f.path
}

func (f *File) Hash() plumbing.Hash {
	return plumbing.ZeroHash
}

func (f *File) Mode() filemode.FileMode {
	return 0
}
