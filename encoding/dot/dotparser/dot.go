// Package dot provides access to Graphviz DOT files.
package dot

import (
	"io"
	"io/ioutil"

	"github.com/graphism/dot/ast"
	"github.com/graphism/dot/internal/lexer"
	"github.com/graphism/dot/internal/parser"
	"github.com/pkg/errors"
)

// ParseFile parses the given Graphviz DOT file into an AST.
func ParseFile(path string) (*ast.File, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ParseBytes(buf)
}

// Parse parses the given Graphviz DOT file into an AST, reading from r.
func Parse(r io.Reader) (*ast.File, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ParseBytes(buf)
}

// ParseBytes parses the given Graphviz DOT file into an AST, reading from b.
func ParseBytes(b []byte) (*ast.File, error) {
	l := lexer.NewLexer(b)
	p := parser.NewParser()
	file, err := p.Parse(l)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	f, ok := file.(*ast.File)
	if !ok {
		return nil, errors.Errorf("invalid file type; expected *ast.File, got %T", file)
	}
	if err := check(f); err != nil {
		return nil, errors.WithStack(err)
	}
	return f, nil
}

// ParseString parses the given Graphviz DOT file into an AST, reading from s.
func ParseString(s string) (*ast.File, error) {
	return ParseBytes([]byte(s))
}
