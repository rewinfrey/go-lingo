package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
)

type Language struct {
	ID         uint `yaml:"language_id"`
	Name       string
	Extensions []string `yaml:"extensions"`
	Filenames  []string `yaml:"filenames"`
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("lingo: ")

	content, err := ioutil.ReadFile("languages.yml")
	if err != nil {
		panic(err)
	}
	languages := map[string]Language{}
	err = yaml.Unmarshal([]byte(content), &languages)
	if err != nil {
		panic(err)
	}

	g := Generator{}
	g.Printf("// Code generated by \"lingo\"; DO NOT EDIT.\n")
	g.Printf("\n")
	g.Printf("package lingo\n")
	g.Printf("\n")
	g.Printf("type Language struct {\n")
	g.Printf("\tID uint\n")
	g.Printf("\tName string\n")
	g.Printf("\tExtensions []string\n")
	g.Printf("\tFilenames []string\n")
	g.Printf("}\n")
	g.Printf("\n")

	g.Printf("var (\n")
	g.Printf("\tLanguages = map[string]Language{\n")
	languagesByExtension := map[string][]Language{}
	languagesByFileName := map[string][]Language{}
	for k, v := range languages {
		g.Printf("\t\t\"%s\": ", k)
		v.Name = k
		g.printLanguage(&v)
		g.Printf(",\n")

		for _, e := range v.Extensions {
			x := languagesByExtension[e]
			x = append(x, v)
			languagesByExtension[e] = x
		}

		for _, e := range v.Filenames {
			x := languagesByFileName[e]
			x = append(x, v)
			languagesByFileName[e] = x
		}
	}
	g.Printf("\t}\n")

	// Languages by extension
	g.Printf("\tLanguagesByExtension = map[string][]Language{\n")
	for ext, langs := range languagesByExtension {
		g.Printf("\t\t\"%s\": []Language{\n", ext)
		for _, l := range langs {
			g.Printf("\t\t\t")
			g.printLanguage(&l)
			g.Printf(",")
			g.Printf("\n")
		}
		g.Printf("\t\t},\n")
	}
	g.Printf("\t}\n")

	// Languages by filename
	g.Printf("\tLanguagesByFileName = map[string][]Language{\n")
	for name, langs := range languagesByFileName {
		g.Printf("\t\t\"%s\": []Language{\n", name)
		for _, l := range langs {
			g.Printf("\t\t\t")
			g.printLanguage(&l)
			g.Printf(",")
			g.Printf("\n")
		}
		g.Printf("\t\t},\n")
	}
	g.Printf("\t}\n")

	// end of `var` declaration
	g.Printf(")\n")

	// Format the output.
	src := g.format()

	err = ioutil.WriteFile("languages.go", src, 0644)
	if err != nil {
		panic(err)
	}
}

func (g *Generator) printLanguage(language *Language) {
	var extensions []string
	for _, e := range language.Extensions {
		extensions = append(extensions, fmt.Sprintf(`"%s"`, e))
	}
	exts := strings.Join(extensions, ", ")
	g.Printf("Language{ID: %d, Name:\"%s\", Extensions: []string{%s} }", language.ID, language.Name, exts)
}

type Generator struct {
	buf bytes.Buffer // Accumulated output.
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}
