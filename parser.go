package dupes

import (
	"crypto/sha1"
	"fmt"
	"hash"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/bblfsh/client-go.v2/tools"
	"gopkg.in/bblfsh/sdk.v1/uast"

	"io/ioutil"

	bblfsh "gopkg.in/bblfsh/client-go.v2"
	"gopkg.in/src-d/enry.v1"
)

type Dupe struct {
	HashStr  string
	LineFrom uint32
	LineTo   uint32
	Filename string
}

type Err struct {
	Filename string
	Err      string
}

type Result struct {
	Dupes [][]*Dupe
	Errs  []*Err
}

type Parser struct {
	client *bblfsh.Client
	hasher hash.Hash
	path   string

	dupes map[string][]*Dupe
	errs  []*Err

	supportedLangs []string
}

func NewParser(c *bblfsh.Client, path string) *Parser {
	return &Parser{
		client: c,
		path:   path,
		hasher: sha1.New(),

		dupes: make(map[string][]*Dupe),
	}
}

func (p *Parser) Parse() (*Result, error) {
	if err := p.fillSupportedLangs(); err != nil {
		return nil, err
	}

	err := filepath.Walk(
		p.path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			return p.parseFile(path)
		},
	)

	if err != nil {
		return nil, err
	}

	return p.fillResultWithDupes(), nil
}

func (p *Parser) fillResultWithDupes() *Result {
	res := &Result{
		Dupes: make([][]*Dupe, 0),
	}

	for _, v := range p.dupes {
		if len(v) > 1 {
			res.Dupes = append(res.Dupes, v)
		}
	}

	res.Errs = p.errs

	return res
}

func (p *Parser) fillSupportedLangs() error {
	if p.supportedLangs == nil {
		res, err := p.client.NewSupportedLanguagesRequest().Do()
		if err != nil {
			return err
		}

		for _, l := range res.Languages {
			p.supportedLangs = append(p.supportedLangs, strings.ToLower(l.Language))
		}
	}

	return nil
}

func (p *Parser) parseFile(file string) error {
	log.Println("parsing file", file)

	lang, content, err := p.getLangAndContent(file)
	if err != nil {
		return err
	}

	if !p.isLangSupported(lang) {
		p.errs = append(p.errs,
			&Err{
				Err:      fmt.Sprintf("language not supported: %s", lang),
				Filename: file,
			},
		)

		return nil
	}

	res, err := p.client.NewParseRequest().Language(lang).Content(string(content)).Do()
	if err != nil {
		p.errs = append(p.errs,
			&Err{
				Err:      fmt.Sprintf("parsing error: %s", err.Error()),
				Filename: file,
			},
		)

		return nil
	}

	nodes, err := tools.Filter(res.UAST, "//*[(@roleFunction and not(@roleCall)) and (@startOffset or @endOffset)]")
	if err != nil {
		return err
	}

	for _, n := range nodes {
		h, hstr, err := p.hash(n)
		if err != nil {
			return err
		}
		dupes := p.dupes[h]
		d := &Dupe{
			HashStr:  hstr,
			Filename: file,
		}
		if n.StartPosition != nil {
			d.LineFrom = n.StartPosition.Line
		}

		if n.EndPosition != nil {
			d.LineTo = n.EndPosition.Line
		}

		p.dupes[h] = append(dupes, d)
	}

	return nil
}

func (p *Parser) isLangSupported(lang string) bool {
	for _, l := range p.supportedLangs {
		if l == lang {
			return true
		}
	}

	return false
}

func (p *Parser) getLangAndContent(file string) (string, []byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", nil, err
	}

	lang := enry.GetLanguage(file, b)

	return strings.ToLower(lang), b, nil
}

func (p *Parser) hash(n *uast.Node) (string, string, error) {
	var tokens []string
	for _, r := range n.Roles {
		tokens = append(tokens, r.String())

	}

	for _, ch := range n.Children {
		_, hstr, err := p.hash(ch)
		if err != nil {
			return "", "", err
		}

		tokens = append(tokens, hstr)
	}

	tokensStr := strings.Join(tokens, ":")

	_, err := p.hasher.Write([]byte(tokensStr))
	if err != nil {
		return "", "", err
	}

	h := fmt.Sprintf("%x", p.hasher.Sum(nil))
	p.hasher.Reset()

	return h, tokensStr, nil
}
