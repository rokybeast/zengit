package git

import "strings"

// simple boilerplate for a new project
const ReadmeTemplate = `# New Project

Initialized with [gitty](https://github.com/rokybeast/gitty).
`

// holds data for a template option
type Template struct {
	Name    string
	Content string
}

// TODO: actually write the gitignores
// make a different folder to handle each gitignore entry
// like: git/gitignores/brainfuck.go (with the stuff in it)
var GitIgnores = []Template{
	{Name: "None", Content: ""},
	{Name: "Go", Content: "bin/\nobj/\n*.out\n"},
	{Name: "Node.JS", Content: "node_modules/\n.env\ndist/\n"},
	{Name: "Python", Content: "__pycache__/\n*.py[cod]\n.venv/\n"},
	{Name: "Rust", Content: "target/"},
}

// TODO: actually write the licenses
// make a different folder to handle each license
// like: git/licenses/mit.go (with the ENTIRE license in it)
var Licenses = []Template{
	{Name: "None", Content: ""},
	{Name: "MIT", Content: strings.TrimSpace(`
MIT License
[im lazy to write, ill do it later]

	`)},
	{Name: "Apache 2.0", Content: strings.TrimSpace(`
Apache License
[again, same thing]
	`)},
}
