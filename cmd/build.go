package cmd

import (
	"blog-typ/temp"
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

type BuildStep struct {
	InFmt string
	OutFmt string
	Hook func(string, []byte) ([]byte, error)
}

func MarkdownHook(_ string, data []byte) ([]byte, error) {
	var result bytes.Buffer

	unsafe := html.WithUnsafe()

	md := goldmark.New(
		goldmark.WithRendererOptions(unsafe),
	)

	if err := md.Convert(data, &result); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

type Driver struct {
	destDir string
	Posts []temp.Post
}

func (d *Driver) AddPost(fname string, hash string) error {
	// parse the date and name from fname
	re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})_(.*)$`)
	matches := re.FindStringSubmatch(fname)
	if matches == nil {
		return errors.New("Invalid format for post file name")
	}

	title := matches[2]

	d.Posts = append(d.Posts, temp.Post { Title: title, Date: matches[1], Hash: hash, Component: nil })

	return nil
}

func (d *Driver) Generate() error {
	driverT := `package static

import "blog-typ/temp"

var Posts = []temp.Post{
{{range $val := .}}
{ "{{$val.Date}}", "{{$val.Title}}", "Static{{$val.Hash}}", Static{{$val.Hash}}},{{end}}
}`

	t, err := template.New("static").Parse(driverT)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(d.destDir, "driver.go"))
	if err != nil {
		return err
	}

	if err := t.Execute(f, d.Posts); err != nil {
		return err
	}

	return nil
}

var GlobalDriver Driver = Driver {
	"",
	[]temp.Post{},
}

func TemplHook(fname string, data []byte) ([]byte, error) {
	templT := `package static

{{if .AddImport}}import "blog-typ/components"{{end}}

templ Static{{ .Name }}() {
{{ .Html }}
}`


	t, err := template.New("static").Parse(templT)
	if err != nil {
		return nil, err
	}

	h := sha1.New()
	h.Write(data)
	hash := fmt.Sprintf("%x", h.Sum(nil))

	addImport := bytes.Contains(data, []byte("@components."))

	apply := struct {
		Name string
		Html string
		AddImport bool
	}{Name: hash, Html: string(data), AddImport: addImport}

	var result bytes.Buffer

	if err := t.Execute(&result, apply); err != nil {
		return nil, err
	}

	if err := GlobalDriver.AddPost(fname, hash); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func BuildItem(fname string, src []byte, steps ...BuildStep) ([]byte, error) {
	var err error
	dest := src

	// Get the initial
	dest, err = steps[0].Hook(fname, dest)
	currFmt := steps[0].OutFmt

	for _i, step := range steps[1:] {
		i := _i + 1

		if currFmt != step.InFmt { // Incompatible build chain
			return nil, errors.New("Incompatible build chain: steps " + strconv.Itoa(i - 1) + " -> " + strconv.Itoa(i) + " step requiring '" + step.InFmt + "' format, received '" + currFmt + "'")
		}

		dest, err = step.Hook(fname, dest)

		if err != nil {
			return nil, err
		}

		currFmt = step.OutFmt
	}

	return dest, nil
}

type BuildCmd struct {
	distPath string
	items map[string][]byte
	steps []BuildStep
}

type Commit struct {
	path string
	data []byte
}

func (c Commit) Exec() error {
	f, err := os.Create(c.path)
	if err != nil {
		return err
	}

	_, err = f.Write(c.data)

	return err
}

type BuildCmdResult struct {
	commits []Commit
}

func (r *BuildCmdResult) Commit() error {
	for _, cm := range r.commits {
		if err := cm.Exec(); err != nil {
			return err
		}
	}

	if err := GlobalDriver.Generate(); err != nil {
		return err
	}

	return nil
}

func (cmd *BuildCmd) Run(args ...string) (CmdResult, error) {
	var result BuildCmdResult

	for name, data := range cmd.items {
		if res, err := BuildItem(name, data, cmd.steps...); err != nil {
			return nil, err
		} else {
			result.commits = append(result.commits, Commit {
				path: filepath.Join(cmd.distPath, name + cmd.steps[len(cmd.steps) - 1].OutFmt),
				data: res,
			})
		}
	}

	return &result, nil
}

// Returns file path/name with the path part and extension part
func fparts(fpath string) (string, string) {
	ext := filepath.Ext(fpath)
	ppart := strings.Replace(fpath, ext, "", 1)

	return ppart, ext
}

// Do nothing Hook
func IdentityHook(_ string, data []byte) ([]byte, error) { return data, nil }

// Note this build command does not recursively check directories
func NewBuildCmd(buildPath string, distPath string) (*BuildCmd, error) {
	buildPath = path.Clean(buildPath)

	GlobalDriver.destDir = distPath

	c := BuildCmd {
		items: make(map[string][]byte),
		steps: []BuildStep{
			{ ".md", ".html", MarkdownHook },
			{ ".html", ".templ", TemplHook },
		},
	}

	entries, err := os.ReadDir(buildPath)
	if err != nil {
		return nil, err
	}

	for i, entry := range entries {
		if entry.IsDir() { continue }

		ppath, ext := fparts(entry.Name())

		// Get the file extension and add a initial "do nothing" hook to intialize the build chain with start constraints
		if firstEntry := i == 0; firstEntry {
			c.steps = append([]BuildStep {{ ext, ext, IdentityHook }}, c.steps...)
		}

		fpath := filepath.Join(buildPath, ppath + ext)
		data, err := os.ReadFile(fpath)
		if err != nil {
			return nil, err
		}

		c.items[ppath] = data
	}

	c.distPath = path.Clean(distPath)

	return &c, nil
}

func (c *BuildCmd) AddStep(hook BuildStep) *BuildCmd {
	c.steps = append(c.steps, hook)
	return c
}

// func main() {
// 	cmd, err := NewBuildCmd("./blogs/", "./dist/")
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	cmd.AddStep(MarkdownHook)
//
// 	result, err := cmd.Run()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	for _, commit := range result.commits {
// 		err := commit.Exec()
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// }
