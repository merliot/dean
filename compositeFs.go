//go:build !tinygo

package dean

import (
	"html/template"
	"io/fs"
	"io/ioutil"
)

// CompositeFS is an ordered (layered) file system, built up from individual
// file systems
type CompositeFS struct {
	fileSystems []fs.ReadFileFS
}

func NewCompositeFS() *CompositeFS {
	return &CompositeFS{}
}

// AddFS adds fsys to the composite fs.  Order matters: first added is lowest
// in priority when searching for a file name in the composite fs.
func (c *CompositeFS) AddFS(fsys fs.ReadFileFS) {
	c.fileSystems = append(c.fileSystems, fsys)
}

func (c *CompositeFS) newest(name string) (fs.File, error) {

	// Start with newest (last added) FS, giving newer FSes priority over
	// older FSes when searching for file name.  The first FS with a
	// matching file name wins.

	for i := len(c.fileSystems) - 1; i >= 0; i-- {
		fsys := c.fileSystems[i]
		if file, err := fsys.Open(name); err == nil {
			return file, nil
		}
	}

	return nil, fs.ErrNotExist
}

// Open a file by name
func (c *CompositeFS) Open(name string) (fs.File, error) {
	return c.newest(name)
}

// Read a file
func (c *CompositeFS) ReadFile(name string) ([]byte, error) {
	file, err := c.newest(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

// ParseFS returns a template by parsing the composite file system for the
// template name matching the pattern name
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
