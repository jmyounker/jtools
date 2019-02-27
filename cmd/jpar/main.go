package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"

	"github.com/jmyounker/jtools/internal/mustache"
)

var version string
var Debug bool = false

const OUTCOME_SUCCESS string = "SUCCESS"
const OUTCOME_FAILURE string = "FAILURE"

func main() {
	err := NewApp().Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type App struct {
	Prog        string
	Parallelism int
	Args        []string
	Dir         string
	Env         map[string]string
	Stdin		string
}

const DEFAULT_PARALLELISM = 8

func NewApp() *App {
	return &App{
		Env:         map[string]string{},
		Parallelism: DEFAULT_PARALLELISM,
		Stdin:	     "{{stdout}}",
	}
}

func (a *App) Run(argv []string) error {
	envPtrn := regexp.MustCompile("^([^=]+)=(.+)$")
	args := []string{}
	a.Prog = argv[0]
	i := 1
	for i < len(argv) {
		x := argv[i]
		switch x {
		case "-p", "--parallelism":
			i = i + 1
			p, err := strconv.Atoi(argv[i])
			if err != nil {
				return err
			}
			a.Parallelism = p
			i = i + 1
		case "-d", "--debug":
			i = i + 1
			Debug = true
		case "-v", "--version":
			i = i + 1
			fmt.Println(version)
			return nil
		case "-h", "--help":
			i = i + 1
			fmt.Printf("usage: %s [--parallelism N] [--debug] CMD\n", a.Prog)
			return nil
		case "--dir":
			i = i + 1
			a.Dir = argv[i]
			i = i + 1
		case "-e", "--env":
			i = i + 1
			matches := envPtrn.FindStringSubmatch(argv[i])
			if len(matches) == 0 {
				return fmt.Errorf("environment variables must have the format var=value and not: %s", argv[i])
			}
			a.Env[matches[1]] = matches[2]
			i = i + 1
		case "-i", "--stdin":
			i = i + 1
			a.Stdin = argv[i]
			i = i + 1
		default:
			// The first unmatched argument to the end of argv is the the whole
			// argument list.
			for i < len(argv) {
				args = append(args, argv[i])
				i = i + 1
			}
		}
	}
	a.Args = args
	return ActionCmd(a)
}

const RETURNCODE_FAILURE = -4242

type Params struct {
	Cmd []*mustache.Template
	Env map[*mustache.Template]*mustache.Template
	Dir *mustache.Template
	Stdin *mustache.Template
}

func ActionCmd(a *App) error {
	params, err := paramsFromApp(a)
	if err != nil {
		return err
	}
	jobs := make(chan Job)
	results := make(chan Output)
	inputDone := make(chan struct{})
	workerDone := make(chan struct{})
	outputDone := make(chan struct{})
	// Launch workers
	for i := 0; i < a.Parallelism; i++ {
		go worker(i, params, jobs, results, workerDone)
	}
	// Send results to workers.
	go func() {
		// Feed input to workers
		j := ReadJsonStream(os.Stdin)
		for x := range j {
			if x.Err == nil {
				jobs <- Job{Value: x.Value}
			} else {
				r := NewJobRun(&[]string{}, "")
				r.Errors = append(r.Errors, fmt.Sprintf("parse error: string(x.Err)"))
				results <- Output{Value: r}
			}
		}
		inputDone <- struct{}{}
	}()
	// Wait for input to complete.
	go func() {
		for x := range results {
			if x.Done {
				break
			} else {
				if !Debug {
					x.Value.Expansions = nil
					x.Value.Stdin = ""
				}
				out, err := json.Marshal(x.Value)
				if err != nil {
					log.Panicf("Cannot marshal internal job record.")
				}
				os.Stdout.Write(out)
			}
		}
		outputDone <- struct{}{}
	}()
	waitForTermination(inputDone, 1)
	// Tell workers that there is no more work.  Workers will
	// now quit.
	for i := 0; i < a.Parallelism; i++ {
		jobs <- Job{Done: true}
	}
	// Wait for workers to complete their current tasks.
	waitForTermination(workerDone, a.Parallelism)
	// Tell output routine that there is nothing left. Output
	// routine will now quit.
	results <- Output{Done: true}
	return nil
}

func paramsFromApp(a *App) (*Params, error) {
	if a.Parallelism < 1 {
		return nil, errors.New("at least one worker required")
	}
	cmd := []*mustache.Template{}
	for _, arg := range a.Args {
		t, err := mustache.ParseString(arg)
		if err != nil {
			return nil, err
		}
		cmd = append(cmd, t)
	}

	env := map[*mustache.Template]*mustache.Template{}
	for k, v := range a.Env {
		kt, err := mustache.ParseString(k)
		if err != nil {
			return nil, fmt.Errorf("cannot parse key %s", k)
		}
		kv, err := mustache.ParseString(v)
		if err != nil {
			return nil, fmt.Errorf("cannot parse value %s", v)
		}
		env[kt] = kv
	}

	dir, err := mustache.ParseString(a.Dir)
	if err != nil {
		return nil, fmt.Errorf("cannot parse dir name: %s", a.Dir)
	}

	stdin, err := mustache.ParseString(a.Stdin)
	if err != nil {
		return nil, fmt.Errorf("cannot parse stdin: %s", a.Stdin)
	}

	return &Params{cmd, env, dir, stdin}, nil
}

func waitForTermination(done chan struct{}, count int) {
	completed := 0
	for range done {
		completed = completed + 1
		if completed == count {
			return
		}
	}
}

func worker(
	id int,
	p *Params,
	jobs chan Job,
	completed chan Output,
	done chan struct{}) {
	for job := range jobs {
		if job.Done {
			done <- struct{}{}
			return
		}
		r := buildJobRun(p, job.Value)
		r = runJob(r)
		if Debug {
			r.WorkerId = &id
		}
		completed <- Output{Value: r}
	}
}

type JobRun struct {
	Cmd        *[]string          `json:"cmd"`
	Prog       *string            `json:"prog"`
	Env        *map[string]string `json:"env,omitempty"`
	Dir        string             `json:"dir,omitempty"`
	Expansions interface{}        `json:"e,omitempty"`
	Returncode int                `json:"returncode"`
	Stdin	   string          	  `json:"stdin,omitempty"`
	Stdout     string             `json:"stdout"`
	Stderr     string             `json:"stderr"`
	Errors     []string           `json:"errors,omitempty"`
	Outcome    string             `json:"outcome"`
	WorkerId   *int               `json:"worker-id,omitempty"`
}

func NewJobRun(cmd *[]string, e interface{}) *JobRun {
	return &JobRun{
		Cmd:        cmd,
		Env:        nil,
		Dir:        "",
		Expansions: e,
		Returncode: RETURNCODE_FAILURE,
		Stdin:      "",
		Stdout:     "",
		Stderr:     "",
		Errors:     []string{},
		Outcome:    OUTCOME_FAILURE,
	}
}

func buildJobRun(params *Params, data interface{}) *JobRun {
	cmd := []string{}
	for _, arg := range params.Cmd {
		cmd = append(cmd, arg.Render(false, data))
	}
	r := NewJobRun(&cmd, data)
	if len(params.Env) > 0 {
		env := map[string]string{}
		for kt, vt := range params.Env {
			k := kt.Render(false, data)
			v := vt.Render(false, data)
			_, ok := env[k]
			if ok {
				r.Errors = append(
					r.Errors,
					fmt.Sprintf("parameter %s is duplicate", k))
				break
			}
			env[k] = v
		}
		r.Env = &env
	}
	if params.Dir != nil {
		r.Dir = params.Dir.Render(false, data)
	}

	r.Stdin = params.Stdin.Render(false, data)

	if len(r.Errors) != 0 {
		r.Outcome = OUTCOME_FAILURE
	} else {
		// Building the job was successful.
		r.Outcome = OUTCOME_SUCCESS
	}
	return r
}

func runJob(r *JobRun) *JobRun {
	// If the outcome is already failure then there is nothing to do.
	if r.Outcome == OUTCOME_FAILURE {
		return r
	}
	cmd0 := (*r.Cmd)[0]
	prog, err := exec.LookPath(cmd0)
	if err != nil {
		r.Outcome = OUTCOME_FAILURE
		r.Errors = append(r.Errors, fmt.Sprintf("cannot locate command %s: %s", cmd0, err))
		return r
	}
	r.Prog = &prog
	c := exec.Cmd{
		Path: prog,
		Args: *r.Cmd,
	}
	if r.Env != nil {
		e := []string{}
		for k, v := range *r.Env {
			e = append(e, fmt.Sprintf("%s=%s", k, v))
		}
		c.Env = e
	}
	if r.Dir != "" {
		c.Dir = r.Dir
	}
	outRdr, err := c.StdoutPipe()
	if err != nil {
		r.Outcome = OUTCOME_FAILURE
		r.Errors = append(r.Errors, fmt.Sprintf("cannot construct stdout: %s", err))
	}
	errRdr, err := c.StderrPipe()
	if err != nil {
		r.Outcome = OUTCOME_FAILURE
		r.Errors = append(r.Errors, fmt.Sprintf("cannot construct stderr: %s", err))
	}
	stdin, err := c.StdinPipe()
	if err != nil {
		r.Outcome = OUTCOME_FAILURE
		r.Errors = append(r.Errors, fmt.Sprintf("cannot construct stdin: %s", err))
	}
	if len(r.Errors) > 0 {
		return r
	}
	err = c.Start()
	if err != nil {
		r.Outcome = OUTCOME_FAILURE
		r.Errors = append(r.Errors, fmt.Sprintf("failed to launch cmd: %s", err))
		return r
	}
	stdout := make(chan StringWithError)
	stderr := make(chan StringWithError)
	go func() {
		stdin.Write([]byte(r.Stdin))
		stdin.Close()
	}()
	go func() {
		out, err := ioutil.ReadAll(outRdr)
		stdout <- StringWithError{string(out), err}
		close(stdout)
	}()
	go func() {
		out, err := ioutil.ReadAll(errRdr)
		stderr <- StringWithError{string(out), err}
		close(stderr)
	}()
	sout := <-stdout
	serr := <-stderr
	r.Stdout = sout.Value
	r.Stderr = serr.Value
	if sout.Err != nil {
		r.Outcome = OUTCOME_FAILURE
		r.Errors = append(r.Errors, fmt.Sprintf("stdout: %s", sout.Err.Error()))
	}
	if serr.Err != nil {
		r.Outcome = OUTCOME_FAILURE
		r.Errors = append(r.Errors, fmt.Sprintf("stderr: %s", serr.Err.Error()))
	}
	c.Wait()
	stat := c.ProcessState.Sys().(syscall.WaitStatus)
	r.Returncode = int(uint32(stat))
	if len(r.Errors) == 0 {
		r.Outcome = OUTCOME_SUCCESS
	}
	return r
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

type StringWithError struct {
	Value string
	Err   error
}

type Job struct {
	Value interface{}
	Done  bool
}

type Output struct {
	Value *JobRun
	Done  bool
}
