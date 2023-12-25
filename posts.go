package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/a-h/templ"
	"github.com/gosimple/slug"
	"github.com/yuin/goldmark"
)

type Post struct {
	Date time.Time	
	Title string
	Content string
}

func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}

func (post *Post) GetSlug() string {
	return path.Join(post.Date.Format("2006/01/02"), slug.Make(post.Title), "/")
}

func NewPost(fileName string) (*Post, error) {
	re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})_(.*)\.md$`)
	matches := re.FindStringSubmatch(fileName)
	if matches == nil {
		return nil, errors.New("Invalid format for post file name")
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	date, err := time.Parse("2006-01-02", matches[1])
	if err != nil {
		return nil, err
	}

	title := matches[2]

	// Convert the content markdown to html
	var buf bytes.Buffer
	if err := goldmark.Convert(data, &buf); err != nil {
		return nil, err
	}

	post := Post {
		Date: date,
		Title: title,
		Content: buf.String(),
	}

	content := Unsafe(buf.String())
	http.Handle("/" + post.GetSlug(), templ.Handler(content))

	return &post, nil
}
