package cmd

import (
	"blog-typ/components"
	static "blog-typ/dist"
	"blog-typ/post"
	"blog-typ/server"
	"bytes"
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/a-h/templ"
	"github.com/yuin/goldmark"
)

type RunCmd struct {
	path string
}

func NewRunCmd(path string) *RunCmd {
	return &RunCmd { path }
}


func MarkdownPlugin(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert(data, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GeneratePosts(path string) ([]*post.Post, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var posts []*post.Post
	for _, entry := range entries {
		if !entry.IsDir() {
			fullPath := filepath.Join(path, entry.Name())

			post, err := post.NewPost(fullPath)
			if err != nil {
				return nil, err
			}
			posts = append(posts, post)
		}
	}

	if len(posts) < 1 {
		return nil, errors.New("No posts found in " + path)
	}

	return posts, nil
}

func (r RunCmd) Run(args ...string) (CmdResult, error) {
	// posts, err := GeneratePosts(r.path)
	// if err != nil {
	// 	slog.Error("Generating posts", "error", err)
	// 	os.Exit(1)
	// }

	for _, post := range static.Posts {
		http.Handle("/" + post.GetSlug(), templ.Handler(post.Component()))
	}

	blog_site := components.Blog("Third Year Project", static.Posts)
	server.Run(blog_site)

	return nil, nil
}
