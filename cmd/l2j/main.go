package main

import (
	"encoding/json"
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/urfave/cli"
)

var version string

func main() {
	app := cli.NewApp()
	app.Usage = "Convert lines to strings."
	app.Action = ActionLinesToStrings
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "chunk-size, stride, c",
			Usage: "Chunk size.",
			Value: 0,
		},
		cli.StringFlag{
			Name:  "split-before",
			Usage: "Split before this regular expression.",
			Value: "",
		},
		cli.StringFlag{
			Name:  "split-after",
			Usage: "Split after this regular expression.",
			Value: "",
		},
	}
	app.Version = version

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ActionLinesToStrings(c *cli.Context) error {
	stride := c.Int("chunk-size")

	split_before, err := CompilePtrn(c.String("split-before"))
	if err != nil {
		return err
	}

	split_after, err := CompilePtrn(c.String("split-after"))
	if err != nil {
		return err
	}

	rd := bufio.NewReader(os.Stdin)
	i := 0
	acc := ""
	var r StringOrError
	for r = range(ReadStrings('\n', rd)) {
		if r.Err != nil {
			break
		}
		i, acc, err = ProcessLine(split_before, split_after, stride, i, acc, r.Value)
		if err != nil {
			return err
		}
	}
	if acc != "" {
		if err := Flush(acc); err != nil {
			return err
		}
	}
	if r.Err != io.EOF {
		return r.Err
	}
	return nil
}

func CompilePtrn(ptrn string) (*regexp.Regexp, error) {
	if ptrn == "" {
		return nil, nil
	} else {
	    return regexp.Compile(ptrn)
	}
}

func ReadStrings(split byte, reader *bufio.Reader) chan StringOrError {
	out := make(chan StringOrError)
	go func() {
		for {
			s, err := reader.ReadString(split)
			if s != "" {
				 out <- StringOrError{s, nil}
			}
			if err != nil {
				out <- StringOrError{"", err}
				close(out)
				return
			}
		}
	}()
	return out
}

type StringOrError struct {
	Value string
	Err error
}

func ProcessLine(split_before, split_after *regexp.Regexp, stride, i int, acc, line string) (int, string, error) {
	if Matches(split_before, line) {
		if acc == "" {
			return 1, line, nil
		}
		if err := Flush(acc); err != nil {
			return 0, "", err
		}
		return 1, line, nil
	} else if stride > 0 && i == stride && Matches(split_after, line) {
		if err := Flush(acc); err != nil {
			return 0, "", err
		}
		if err := Flush(line); err != nil {
			return 0, "", err
		}
		return 0, "", nil
	} else if stride > 0 && i == stride {
		if err := Flush(acc); err != nil {
			return 0, "", err
		}
		return 1, line, nil
	} else if Matches(split_after, line) {
		if err := Flush(acc + line); err != nil {
			return 0, "", err
		}
		return 0, "", nil
	} else {
		return i + 1, acc + line, nil
	}
}


func Matches(r *regexp.Regexp, s string) bool {
	if r == nil {
		return false
	}
	return r.MatchString(s)
}

func Flush(s string) error {
	out, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

