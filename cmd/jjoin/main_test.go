package main

import (
	"testing"
	"os"
	"io/ioutil"
	"bytes"
	"log"
	"github.com/urfave/cli"
	"encoding/json"
	"fmt"
)

func failWhenErr(t *testing.T, err error) {
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func failWhen(t *testing.T, x bool) {
	if x {
		t.Fail()
	}
}


type TestCase struct {
	Left string
	Right string
	Join string
	LeftKey string
	RightKey string
	Pairs []JoinedPair
}

var joinTests = []TestCase{
	{
		"{\"a\": \"b\"}",
		"\"b\"",
		"inner",
		"{{a}}",
		"{{.}}",
		[]JoinedPair{
			{Left: map[string]string{"a": "b"},
			Right:"b"},
		},
	},
	{
		"{\"a\": \"b\"}",
		"{\"a\": \"b\"}",
		"inner",
		"{{a}}",
		"{{a}}",
		[]JoinedPair{
			{Left: map[string]string{"a": "b"},
			 Right: map[string]string{"a": "b"}},
		},
	},
}

func TestJoin(t *testing.T) {
	for _, tc := range joinTests {
		left, err := ioutil.TempFile("", "left.json")
		failWhenErr(t, err)
		defer os.Remove(left.Name())
		left.WriteString(tc.Left)
		left.Close()

		right, err := ioutil.TempFile("", "right.json")
		failWhenErr(t, err)
		defer os.Remove(right.Name())
		right.WriteString(tc.Right)
		right.Close()

		out, err := ioutil.TempFile("", "output")
		failWhenErr(t, err)
		defer os.Remove(out.Name())
		out.Close()

		cmd := []string{
			"jjoin",
			tc.Join,
			"--left", left.Name(),
			"--right", right.Name(),
			"--left-key", tc.LeftKey,
			"--right-key", tc.RightKey,
			"--output", out.Name(),
		}
		err = buildApp().Run(cmd)
		stdout, err := ioutil.ReadFile(out.Name())
		failWhenErr(t, err)
		res := JoinedPair{}
		err = json.Unmarshal(stdout, &res)
		failWhenErr(t, err)
		if fmt.Sprintf("%s", tc.Pairs[0]) != fmt.Sprintf("%s", res) {
			log.Printf("Want: %s", tc.Pairs[0])
			log.Printf("Got: %s", res)
			t.Fail()
		}
	}
}

func RunApp(app *cli.App, cmd []string) (error, string) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	err := app.Run(cmd)
	log.SetOutput(os.Stderr)
	return err, buf.String()
}

