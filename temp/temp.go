package temp

import (
	"path/filepath"

	"github.com/a-h/templ"
	"github.com/gosimple/slug"
)

func (p *Post) GetSlug() string {
	return filepath.Join(p.Date, slug.Make(p.Title), "/")
}

type ComponentFunc func() templ.Component

type Post struct {
	Date string
	Title string
	Hash string
	Component ComponentFunc
}
