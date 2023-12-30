package post

import (
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
)

type PluginFunc func ([]byte) ([]byte, error)

type Post struct {
	Date time.Time	
	Title string
	Content []byte
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

func NewPost(fileName string, plugins ...PluginFunc) (*Post, error) {
	re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})_(.*)\.html$`)
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

	for _, plugin := range plugins {
		data, err = plugin(data)
		if err != nil {
			return nil, err
		}
	}

	post := Post {
		Date: date,
		Title: title,
		Content: data,
	}

	content := Unsafe(string(data))
	http.Handle("/" + post.GetSlug(), templ.Handler(content))

	return &post, nil
}
