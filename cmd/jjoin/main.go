package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli"
	"github.com/jmyounker/mustache"
)

var version string

func main() {

	err := buildApp().Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func buildApp() *cli.App {
	app := cli.NewApp()
	app.Usage = "Join dictionaries from JSON streams."
	app.Version = version

	stdFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "left, l",
			Usage: "Left JSON stream.",
		},
		cli.StringFlag{
			Name:  "right, r",
			Usage: "Right JSON stream.",
		},
		cli.StringFlag{
			Name:  "left-key, lk",
			Usage: "Left stream join key.",
		},
		cli.StringFlag{
			Name:  "right-key, rk",
			Usage: "Right stream join key.",
		},
		cli.StringFlag{
			Name:  "using, u",
			Usage: "Join both streams using this key.",
		},
		cli.StringFlag{
			Name:  "output, o",
			Usage: "Send output to file instead of stdout.",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "inner",
			Usage:  "Inner join",
			Action: actionInnerJoin,
			Flags:  stdFlags,
		},
		{
			Name:   "full",
			Usage:  "Full outer join",
			Action: actionFullOuterJoin,
			Flags:  stdFlags,
		},
		{
			Name:   "outer",
			Usage:  "Full outer join",
			Action: actionFullOuterJoin,
			Flags:  stdFlags,
		},
		{
			Name:   "outer-full",
			Usage:  "Full outer join",
			Action: actionFullOuterJoin,
			Flags:  stdFlags,
		},
		{
			Name:   "full-outer",
			Usage:  "Full outer join",
			Action: actionFullOuterJoin,
			Flags:  stdFlags,
		},
		{
			Name:   "left",
			Usage:  "Left outer join",
			Action: actionLeftOuterJoin,
			Flags:  stdFlags,
		},
		{
			Name:   "left-outer",
			Usage:  "Left outer join",
			Action: actionLeftOuterJoin,
			Flags:  stdFlags,
		},
		{
			Name:   "outer-left",
			Usage:  "Left outer join",
			Action: actionLeftOuterJoin,
			Flags:  stdFlags,
		},
		{
			Name:   "right",
			Usage:  "Right outer join",
			Action: actionRightOuterJoin,
			Flags:  stdFlags,
		},
		{
			Name:   "right-outer",
			Usage:  "Right outer join",
			Action: actionRightOuterJoin,
			Flags:  stdFlags,
		},
		{
			Name:   "outer-right",
			Usage:  "Right outer join",
			Action: actionRightOuterJoin,
			Flags:  stdFlags,
		},
		{
			Name:   "symm-diff",
			Usage:  "Symmetric difference",
			Action: actionSymmetricDiff,
			Flags:  stdFlags,
		},
		{
			Name:   "subtract",
			Action: actionSubtract,
			Usage:  "Subtract right stream from left stream.",
			Flags:  stdFlags,
		},
	}
	return app
}

func actionInnerJoin(c *cli.Context) error {
	return genericJoinAction(c, true, false, false)
}

func actionFullOuterJoin(c *cli.Context) error {
	return genericJoinAction(c, true, true, true)
}

func actionLeftOuterJoin(c *cli.Context) error {
	return genericJoinAction(c, true, true, false)
}

func actionRightOuterJoin(c *cli.Context) error {
	return genericJoinAction(c, true, false, true)
}

func actionSymmetricDiff(c *cli.Context) error {
	return genericJoinAction(c, false, true, true)
}

func genericJoinAction(c *cli.Context, inner, left, right bool) error {
	params, err := PopulateJoin(c)
	if err != nil {
		return err
	}
	joined := PerformJoin(params, inner, left, right)
	out, err := GetOutputFile(c)
	if err != nil {
		return err
	}
	defer out.Close()
	DisplayJoinedPairs(joined, out)
	return nil
}

func actionSubtract(c *cli.Context) error {
	params, err := PopulateJoin(c)
	if err != nil {
		return err
	}
	joined := PerformJoin(params, false, true, false)
	out, err := GetOutputFile(c)
	if err != nil {
		return err
	}
	defer out.Close()
	for _, x := range joined {
		b, err := json.MarshalIndent(x.Left, "", "  ")
		if err != nil {
			panic("error rendering JSON")
		}
		fmt.Fprint(out, string(b))
	}
	return nil
}

type JoinParams struct {
	Left     []interface{}
	Right    []interface{}
	LeftKey  *Key
	RightKey *Key
}

func PopulateJoin(c *cli.Context) (*JoinParams, error) {
	ls, err := getDataStream("left", c)
	if err != nil {
		return nil, err
	}
	rs, err := getDataStream("right", c)
	if err != nil {
		return nil, err
	}
	lk, rk, err := getKeyOpts(c)
	if err != nil {
		return nil, err
	}
	return &JoinParams{
		Left:     ls,
		Right:    rs,
		LeftKey:  lk,
		RightKey: rk,
	}, nil
}

func getDataStream(streamOpt string, c *cli.Context) ([]interface{}, error) {
	fn := c.String(streamOpt)
	if fn == "" {
		return nil, fmt.Errorf("%s data stream required", streamOpt)
	}
	f, err := os.OpenFile(c.String(streamOpt), os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s", fn)
	}
	return decodeJsonStream(f)
}

func getKeyOpts(c *cli.Context) (*Key, *Key, error) {
	leftKeyExpr := c.String("left-key")
	rightKeyExpr := c.String("right-key")
	bothKeyExpr := c.String("using")

	if bothKeyExpr == "" && leftKeyExpr == "" && rightKeyExpr == "" {
		return nil, nil, errors.New("keys required")
	}
	if bothKeyExpr != "" {
		if leftKeyExpr != "" || rightKeyExpr != "" {
			return nil, nil, errors.New("using is mutually exclusive with left and right key")
		}
		k, err := MustacheKeyFromString(bothKeyExpr)
		if err != nil {
			return nil, nil, err
		}
		return k, k, nil
	}
	if leftKeyExpr == "" || rightKeyExpr == "" {
		return nil, nil, errors.New("both left key and right key are required")
	}
	kl, err := MustacheKeyFromString(leftKeyExpr)
	if err != nil {
		return nil, nil, fmt.Errorf("left key error: %s", err)
	}
	kr, err := MustacheKeyFromString(rightKeyExpr)
	if err != nil {
		return nil, nil, fmt.Errorf("right key error: %s", err)
	}
	return kl, kr, nil
}

func decodeJsonStream(in *os.File) ([]interface{}, error) {
	dec := json.NewDecoder(in)
	r := []interface{}{}
	for {
		var j interface{}
		if err := dec.Decode(&j); err != nil {
			if err == io.EOF {
				return r, nil
			}
			return nil, err
		}
		r = append(r, j)
	}
	return r, nil
}

func PerformJoin(p *JoinParams, inner, leftOuter, rightOuter bool) []JoinedPair {
	leftByKey := PartitionByKey(p.LeftKey.Get, p.Left)
	rightByKey := PartitionByKey(p.RightKey.Get, p.Right)
	keys := UnionKeys(leftByKey, rightByKey)
	j := []JoinedPair{}
	for k := range keys {
		left, ok := leftByKey[k]
		if !ok {
			left = []interface{}{}
		}
		right, ok := rightByKey[k]
		if !ok {
			right = []interface{}{}
		}
		if inner && len(left) > 0 && len(right) > 0 {
			for _, xl := range left {
				for _, xr := range right {
					j = append(j, JoinedPair{xl, xr})
				}
			}
		}
		if leftOuter && len(left) > 0 && len(right) == 0 {
			for _, x := range left {
				j = append(j, JoinedPair{x, nil})
			}

		}
		if rightOuter && len(left) == 0 && len(right) > 0 {
			for _, x := range right {
				j = append(j, JoinedPair{nil, x})
			}
		}
	}
	return j
}

type JoinedPair struct {
	Left  interface{}
	Right interface{}
}

func DisplayJoinedPairs(p []JoinedPair, f *os.File) {
	for _, x := range p {
		b, err := json.Marshal(
			map[string]interface{}{
				"left":  x.Left,
				"right": x.Right,
			})
		if err != nil {
			panic("error rendering JSON")
		}
		fmt.Fprintf(f, string(b))
	}
}

func GetOutputFile(c *cli.Context) (*os.File, error) {
	fn := c.String("output")
	if fn == "" {
		return os.Stdout, nil
	}
	f, err := os.Create(fn)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func PartitionByKey(key func(interface{}) string, seq []interface{}) map[interface{}][]interface{} {
	part := map[interface{}][]interface{}{}
	for _, x := range seq {
		ks := key(x)
		pk, ok := part[ks]
		if !ok {
			pk = []interface{}{}
		}
		part[ks] = append(pk, x)
	}
	return part
}

func UnionKeys(left map[interface{}][]interface{}, right map[interface{}][]interface{}) map[interface{}]struct{} {
	k := map[interface{}]struct{}{}
	for kl := range left {
		k[kl] = struct{}{}
	}
	for kr := range right {
		k[kr] = struct{}{}
	}
	return k
}

type Key struct {
	tmpl *mustache.Template
}

func MustacheKeyFromString(s string) (*Key, error) {
	tmpl, err := mustache.ParseString(s)
	if err != nil {
		return nil, err
	}
	return &Key{tmpl}, nil
}

func (k Key) Get(ctx interface{}) string {
	return k.tmpl.Render(false, ctx)

}
