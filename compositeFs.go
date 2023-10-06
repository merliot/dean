//go:build !tinygo

package dean

import (
	"io/fs"
	"html/template"
)

type CompositeFS struct {
	fileSystems []fs.FS
}

func NewCompositeFS() *CompositeFS {
	return &CompositeFS{}
}

func (c *CompositeFS) AddFS(fsys fs.FS) {
	c.fileSystems = append(c.fileSystems, fsys)
}

func (c *CompositeFS) Open(name string) (fs.File, error) {

	// Start with newest (last added) FS, giving newer FSes priority over
	// older FSes when searching for file name.  The first FS with a
	// matching file name wins.

	for i := len(c.fileSystems)-1; i >= 0; i-- {
		fsys := c.fileSystems[i]
		if file, err := fsys.Open(name); err == nil {
			return file, nil
		}
	}

	return nil, fs.ErrNotExist
}

func (c *CompositeFS) ParseFS(pattern string) *template.Template {

	// Iterate from oldest (first added) FS to newest FS, building a "main"
	// template with pattern matching templates from each FS.  The winner
	// for when templates have the same name will be the last one added to
	// the main template (newest FS wins). 

	mainTmpl := template.New("main")

	for _, fsys := range c.fileSystems {
		mainTmpl.ParseFS(fsys, pattern)
	}

	return mainTmpl
}
