// resume-gen renders resume.yaml into a PDF using an embedded LaTeX template.
//
//	go run . [-in resume.yaml] [-out resume.pdf] [-tex]
//
// Requires pdflatex in PATH (TeX Live / MacTeX / MiKTeX).
package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:embed template.tex
var texTemplate string

type Resume struct {
	Name       string       `yaml:"name"`
	Contacts   []string     `yaml:"contacts"`
	Education  []Education  `yaml:"education"`
	Languages  []Language   `yaml:"languages"`
	Experience []Experience `yaml:"experience"`
	Skills     []Skill      `yaml:"skills"`
	Projects   []Project    `yaml:"projects"`
}

type Education struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
	Degree   string `yaml:"degree"`
	Dates    string `yaml:"dates"`
}

type Language struct {
	Name  string `yaml:"name"`
	Level string `yaml:"level"`
}

type Experience struct {
	Company  string `yaml:"company"`
	Location string `yaml:"location"`
	Role     string `yaml:"role"`
	Dates    string `yaml:"dates"`
}

type Skill struct {
	Category string   `yaml:"category"`
	Items    []string `yaml:"items"`
}

type Project struct {
	Name             string   `yaml:"name"`
	Dates            string   `yaml:"dates"`
	Stack            []string `yaml:"stack"`
	URL              string   `yaml:"url"`
	Description      string   `yaml:"description"`
	Responsibilities []string `yaml:"responsibilities"`
}

// texEscaper turns plain text into LaTeX-safe text.
var texEscaper = strings.NewReplacer(
	`\`, `\textbackslash{}`,
	`&`, `\&`,
	`%`, `\%`,
	`$`, `\$`,
	`#`, `\#`,
	`_`, `\_`,
	`{`, `\{`,
	`}`, `\}`,
	`^`, `\textasciicircum{}`,
	`~`, `$\sim$`,
	`²`, `\textsuperscript{2}`,
	`³`, `\textsuperscript{3}`,
	`∼`, `$\sim$`,
	`·`, `$\cdot$`,
	`—`, `---`,
	`–`, `--`,
	`’`, `'`,
	`“`, "``",
	`”`, `''`,
	`€`, `EUR`,
)

func tex(s string) string {
	// YAML folded/literal blocks may contain newlines; collapse to single spaces.
	return texEscaper.Replace(strings.Join(strings.Fields(s), " "))
}

func join(items []string) string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = tex(it)
	}
	return strings.Join(out, ", ")
}

func main() {
	in := flag.String("in", "resume.yaml", "input YAML file")
	out := flag.String("out", "resume.pdf", "output PDF file")
	keepTex := flag.Bool("tex", false, "also write the generated .tex next to the PDF")
	flag.Parse()

	if err := run(*in, *out, *keepTex); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(in, out string, keepTex bool) error {
	raw, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	var r Resume
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true) // catch typos in field names
	if err := dec.Decode(&r); err != nil {
		return fmt.Errorf("parse %s: %w", in, err)
	}

	tpl, err := template.New("resume").
		Delims("<<", ">>").
		Funcs(template.FuncMap{"tex": tex, "join": join}).
		Parse(texTemplate)
	if err != nil {
		return fmt.Errorf("template: %w", err)
	}
	var src bytes.Buffer
	if err := tpl.Execute(&src, r); err != nil {
		return fmt.Errorf("render: %w", err)
	}

	if keepTex {
		texPath := strings.TrimSuffix(out, filepath.Ext(out)) + ".tex"
		if err := os.WriteFile(texPath, src.Bytes(), 0o644); err != nil {
			return err
		}
		fmt.Println("wrote", texPath)
	}

	tmp, err := os.MkdirTemp("", "resume-gen-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	if err := os.WriteFile(filepath.Join(tmp, "resume.tex"), src.Bytes(), 0o644); err != nil {
		return err
	}
	cmd := exec.Command("pdflatex", "-interaction=nonstopmode", "-halt-on-error", "resume.tex")
	cmd.Dir = tmp
	logOut, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pdflatex failed: %w\n%s", err, lastLines(string(logOut), 25))
	}

	pdf, err := os.ReadFile(filepath.Join(tmp, "resume.pdf"))
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, pdf, 0o644); err != nil {
		return err
	}
	fmt.Println("wrote", out)
	return nil
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
