package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli"

	"github.com/jmyounker/jtools/internal/mustache"
)

var version string

func main() {
	app := cli.NewApp()
	app.Usage = "Join dictionaries from JSON streams."
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "input, i",
			Usage: "Input filename.",
		},
		cli.StringFlag{
			Name:  "template, t",
			Usage: "Template filename.",
		},
		cli.StringFlag{
			Name:  "template-expansion, tx",
			Usage: "Template filename expansion.",
		},
		cli.StringFlag{
			Name:  "output, o",
			Usage: "Output filename.",
		},
		cli.StringFlag{
			Name:  "output-expansion, ox",
			Usage: "Output filename expansion.",
		},
		cli.BoolFlag{
			Name:  "append, a",
			Usage: "Append to file.",
		},
		cli.BoolFlag{
			Name:  "newline, n",
			Usage: "End each expansion with a newline.",
		},
		cli.BoolFlag{
			Name:  "html, strict-mustache",
			Usage: "Use vanilla mustache expansions in the main template.",
		},
	}
	app.Action = JxAction

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func JxAction(c *cli.Context) {
	in, err := getInput(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(127)
	}

	out, err := getOutputFactory(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(127)
	}

	tmpl, err := getTemplateFactory(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(127)
	}

	if err := expand(in, out, tmpl, c.Bool("newline"), c.Bool("html")); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func getInput(c *cli.Context) (io.Reader, error) {
	inputName := c.String("input")
	if inputName == "" {
		return os.Stdin, nil
	}
	in, err := os.Open(inputName)
	if err != nil {
		return nil, fmt.Errorf("could not open file %q for reading: %s", inputName, err)
	}
	return in, nil
}

func getOutputFactory(c *cli.Context) (writerFactory, error) {
	outputName := c.String("output")
	outputXpn := c.String("output-expansion")
	tmplXpn := c.String("template-expansion")
	append := c.Bool("append")
	if outputName == "" && outputXpn == "" {
		return &staticWriterFactory{os.Stdout}, nil
	}
	if outputName != "" && outputXpn != "" {
		return nil, fmt.Errorf("-o and -ox are mutally exclusive")
	}
	if outputName != "" {
		out, err := openFile(outputName, append)
		if err != nil {
			return nil, fmt.Errorf("could not open file %q for writing: %s", outputName, err)
		}
		return &staticWriterFactory{out}, nil
	}
	fnTmpl, err := mustache.ParseString(outputXpn)
	if err != nil {
		return nil, fmt.Errorf("could not parse output path template %q: %s", tmplXpn, err)
	}
	return &dynamicWriterFactory{fnTmpl: fnTmpl, append: append}, nil
}

func getTemplateFactory(c *cli.Context) (templateFactory, error) {
	tmplName := c.String("template")
	tmplXpn := c.String("template-expansion")
	if tmplName == "" && tmplXpn == "" && c.NArg() == 0 {
		return nil, errors.New("you must supply a template or -t")
	}
	if tmplName != "" && tmplXpn != "" {
		return nil, fmt.Errorf("-t and -tx are mutually exclusive")
	}
	if tmplName == "" && tmplXpn == "" {
		tmplName = c.Args().Get(0)
		tmpl, err := mustache.ParseString(tmplName)
		if err != nil {
			return nil, fmt.Errorf("could not parse template %q: %s", tmplName, err)
		}
		return &staticTemplateFactory{tmpl}, nil
	}

	if tmplName != "" {
		tmpl, err := mustache.ParseFile(tmplName)
		if err != nil {
			return nil, fmt.Errorf("could not parse template file %q: %s", tmplName, err)
		}
		return &staticTemplateFactory{tmpl}, nil
	}

	fnTmpl, err := mustache.ParseString(tmplXpn)
	if err != nil {
		return nil, fmt.Errorf("could not parse template path template %q: %s", tmplXpn, err)
	}
	return &dynamicTemplateFactory{fnTmpl: fnTmpl}, nil
}

// writerFactory allows choosing output sources based on the JSON input
type writerFactory interface {
	getWriter(xpn interface{}) (io.Writer, error)
}

type staticWriterFactory struct {
	writer io.Writer
}

func (f *staticWriterFactory) getWriter(xpn interface{}) (io.Writer, error) {
	return f.writer, nil
}

type dynamicWriterFactory struct {
	fnTmpl         *mustache.Template // filename template
	append         bool               // append to file rather than truncate
	strictMustache bool               // use strict mustache expansion
	fn             string             // path to current template
	writer         *os.File           // current writer
}

func (f *dynamicWriterFactory) getWriter(xpn interface{}) (io.Writer, error) {
	fn := f.fnTmpl.Render(f.strictMustache, xpn)
	if fn == f.fn {
		return f.writer, nil
	}
	if f.writer != nil {
		f.writer.Close()
	}
	writer, err := openFile(fn, f.append)
	if err != nil {
		return nil, err
	}
	f.fn = fn
	f.writer = writer
	return writer, nil
}

// extracted for cleanliness
func openFile(fn string, append bool) (*os.File, error) {
	md := os.O_TRUNC
	if append {
		md = os.O_APPEND
	}
	return os.OpenFile(fn, md|os.O_CREATE|os.O_WRONLY, 0666)
}

// templateFactory allows choosing templates based on the JSON input
type templateFactory interface {
	getTemplate(xpn interface{}) (*mustache.Template, error)
}

type staticTemplateFactory struct {
	tmpl *mustache.Template
}

func (f *staticTemplateFactory) getTemplate(xpn interface{}) (*mustache.Template, error) {
	return f.tmpl, nil
}

type dynamicTemplateFactory struct {
	fnTmpl *mustache.Template // filename template
	fn     string             // path to current template
	tmpl   *mustache.Template // current template
}

func (dtf *dynamicTemplateFactory) getTemplate(xpn interface{}) (*mustache.Template, error) {
	fn := dtf.fnTmpl.Render(false, xpn)
	if fn == dtf.fn {
		return dtf.tmpl, nil
	}
	tmpl, err := mustache.ParseFile(fn)
	if err != nil {
		return nil, err
	}
	dtf.tmpl = tmpl
	return tmpl, nil
}

// expand combines JSON input with templates to produce output.
func expand(in io.Reader, outFact writerFactory, tmplFact templateFactory, newline bool, strictMustache bool) error {
	dec := json.NewDecoder(in)
	var j interface{}
	for {
		if err := dec.Decode(&j); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		tmpl, err := tmplFact.getTemplate(j)
		if err != nil {
			return err
		}

		out, err := outFact.getWriter(j)
		if err != nil {
			return err
		}

		out.Write([]byte(tmpl.Render(strictMustache, j)))
		if newline {
			out.Write([]byte("\n"))
		}
	}
	return nil
}
