package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/require"
)

func TestGetTypeOfChange(t *testing.T) {
	tests := []struct {
		name     string
		commits  []string
		expected string
	}{
		{
			name:     "Empty commits",
			commits:  []string{},
			expected: patch,
		},
		{
			name:     "Feature commit",
			commits:  []string{"feat(something): add login feature"},
			expected: minor,
		},
		{
			name:     "Multiple commits, last is feat",
			commits:  []string{"chore(scope): update dependencies", "feat: add new API endpoint"},
			expected: minor,
		},
		{
			name:     "Multiple commits, last is breaking",
			commits:  []string{"feat: add something", "fix!: critical fix"},
			expected: major,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTypeOfChange(tt.commits)
			if result != tt.expected {
				t.Errorf("getTypeOfChange(%v) = %v, want %v", tt.commits, result, tt.expected)
			}
		})
	}
}

func createCommit(t *testing.T, r *git.Repository, path string, message string) plumbing.Hash {
	// Create a commit so there is a HEAD to check
	wt, err := r.Worktree()
	require.NoError(t, err)

	rm, err := wt.Filesystem.Create(path)

	require.NoError(t, err)

	randomNumber := rand.Intn(10000000000)
	_, err = rm.Write([]byte(fmt.Sprint(randomNumber)))
	require.NoError(t, err)

	_, err = wt.Add(path)
	require.NoError(t, err)

	author := object.Signature{
		Name:  "go-git",
		Email: "go-git@fake.local",
		When:  time.Now(),
	}

	h, err := wt.Commit(message, &git.CommitOptions{
		All:       true,
		Author:    &author,
		Committer: &author,
	})
	require.NoError(t, err)
	return h
}

func TestGit(t *testing.T) {
	fs := memfs.New()
	r, err := git.Init(memory.NewStorage(), fs)
	require.NoError(t, err)
	module := "helpers"
	hash := createCommit(t, r, fmt.Sprintf("%s/foo.go", module), "first commit")
	latestTag := fmt.Sprintf("%s/v0.0.1", module)
	_, err = r.CreateTag(latestTag, hash, nil)
	require.NoError(t, err)
	version, err := semver.NewVersion("0.0.1")
	require.NoError(t, err)
	commitAfterTag := "second commit"
	createCommit(t, r, fmt.Sprintf("%s/foo.go", module), commitAfterTag)

	cm, err := getCommitMessagesFromLastTag(r, version, module)
	require.NoError(t, err)
	require.Equal(t, 1, len(cm))
	require.Equal(t, commitAfterTag, cm[0])

}
