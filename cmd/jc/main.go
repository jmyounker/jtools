package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli"
	"io"
)

var version string

func main() {
	app := cli.NewApp()
	app.Usage = "Count JSON data."
	app.Action = ActionCount
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "statistics, s",
			Usage: "Emit statistics insead of records.",
		},
		cli.IntFlag{
			Name:  "chunk-size, stride, c",
			Usage: "Chunk size.",
			Value: 1,
		},
		cli.StringFlag{
			Name:  "chunk-counter-var",
			Usage: "Chunk counter attribute.",
			Value: "c",
		},
		cli.StringFlag{
			Name:  "counter-var",
			Usage: "Simple counter attribute.",
			Value: "i",
		},
		cli.StringFlag{
			Name:  "stride-var",
			Usage: "Stride counter attribute.",
			Value: "s",
		},
		cli.StringFlag{
			Name:  "body-var",
			Usage: "Body variable",
			Value: "e",
		},
	}
	app.Version = version

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ActionCount(c *cli.Context) error {
	j := ReadJsonStream(os.Stdin)
	reportStatistics := c.Bool("statistics")
	chunkSize := c.Int("chunk-size")
	counter := c.String("counter-var")
	strideCounter := c.String("stride-var")
	chunkCounter := c.String("chunk-counter-var")
	body := c.String("body-var")
	i := 0
	s := 0
	cc := 0
	b := map[string]interface{}{}
	for x := range j {
		if x.Err != nil {
			return x.Err
		}
		if !reportStatistics {
			b[counter] = i
			b[strideCounter] = s
			b[chunkCounter] = cc
			b[body] = x.Value
			out, err := json.MarshalIndent(b, "", "  ")
			if err != nil {
				return err
			}
			fmt.Print(string(out))
		}
		i = i + 1
		s = s + 1
		if s >= chunkSize {
			s = 0
			cc = cc + 1
		}
	}
	if reportStatistics {
		b["records"] = i
		b["chunks"] = cc
		b["chunk-size"] = chunkSize
		out, err := json.MarshalIndent(b, "", "  ")
		if err != nil {
			return err
		}
		fmt.Print(string(out))
	}
	return nil
}

func ReadJsonStream(stream *os.File) chan JsonRead {
	dec := json.NewDecoder(stream)
	out := make(chan JsonRead)
	var j interface{}
	go func() {
		for {
			if err := dec.Decode(&j); err != nil {
				if err == io.EOF {
					close(out)
					return
				} else {
					out <- JsonRead{nil, err}
					close(out)
					return
				}
			}
			out <- JsonRead{j, nil}
		}
	}()
	return out
}

type JsonRead struct {
	Value interface{}
	Err   error
}
