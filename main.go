package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/a-h/templ"
)

type Config struct {
	Posts string
}

func generatePosts(path string) ([]*Post, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var posts []*Post
	for _, entry := range entries {
		if !entry.IsDir() {
			post, err := NewPost("./blogs/" + entry.Name())
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

func main() {
	var conf Config
	_, err := toml.DecodeFile("./config.toml", &conf)
	if err != nil {
		panic(err)
	}

	posts, err := generatePosts(conf.Posts)
	if err != nil {
		panic(err)
	}

	blog_site := blog("Third Year Project", posts)

	http.Handle("/", templ.Handler(blog_site))

	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", nil)
}
