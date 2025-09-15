package git

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommitString(t *testing.T) {
	tests := []struct {
		name   string
		commit Commit
		want   string
	}{
		{
			name:   "empty",
			commit: Commit{},
			want:   "",
		},
		{
			name: "one file",
			commit: Commit{
				Hash: "539533aab24270f6201fcdd5aa25f6c16662ee58",
				Diff: Diff{
					Data: map[string]string{
						"notes.md": "# Notes# Notes\n\n## One more note",
					},
				},
			},
			want: "# Notes# Notes\n\n## One more note\n",
		},
		{
			name: "multiple files",
			commit: Commit{
				Hash: "539533aab24270f6201fcdd5aa25f6c16662ee58",
				Diff: Diff{
					Data: map[string]string{
						"z.md": "# Notes\nSome Note",
						"b.md": "# Iam readme file",
						"a.md": "Just some other file\n\nWith some text",
					},
				},
			},
			want: "Just some other file\n\nWith some text\n# Iam readme file\n# Notes\nSome Note\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commit := tt.commit
			assert.Equalf(t, tt.want, commit.String(), "String()")
		})
	}
}

func TestCommitGetFileNameByLine(t *testing.T) {

	tests := []struct {
		name         string
		lineNum      int
		commit       Commit
		wantFileName string
		wantErr      assert.ErrorAssertionFunc
	}{
		{
			name:         "empty",
			lineNum:      1,
			commit:       Commit{},
			wantFileName: "",
			wantErr:      assert.Error,
		},
		{
			name:    "line 0",
			lineNum: 0,
			commit: Commit{
				Hash: "539533aab24270f6201fcdd5aa25f6c16662ee58",
				Diff: Diff{
					Data: map[string]string{
						"z.md": "# Notes\nSome Note",
						"b.md": "# Iam readme file",
						"a.md": "Just some other file\n\nWith some text",
					},
				},
			},
			wantFileName: "",
			wantErr:      assert.Error,
		},
		{
			name:    "first line of first file",
			lineNum: 1,
			commit: Commit{
				Hash: "539533aab24270f6201fcdd5aa25f6c16662ee58",
				Diff: Diff{
					Data: map[string]string{
						"z.md": "# Notes\nSome Note",
						"b.md": "# Iam readme file",
						"a.md": "Just some other file\n\nWith some text",
					},
				},
			},
			wantFileName: "a.md",
			wantErr:      assert.NoError,
		},
		{
			name:    "one file one line",
			lineNum: 1,
			commit: Commit{
				Hash: "539533aab24270f6201fcdd5aa25f6c16662ee58",
				Diff: Diff{
					Data: map[string]string{
						"test.md": "# Notes",
					},
				},
			},
			wantFileName: "test.md",
			wantErr:      assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFileName, err := tt.commit.GetFileNameByLine(tt.lineNum)
			if !tt.wantErr(t, err, fmt.Sprintf("GetFileNameByLine(%v)", tt.lineNum)) {
				return
			}
			assert.Equalf(t, tt.wantFileName, gotFileName, "GetFileNameByLine(%v)", tt.lineNum)
		})
	}

}
