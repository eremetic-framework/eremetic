package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/klarna/eremetic"
	"github.com/klarna/eremetic/client"
)

var defaultEremeticServer = "http://localhost:8000"

func main() {
	srv := os.Getenv("EREMETIC_URL")
	if srv == "" {
		srv = defaultEremeticServer
	}

	httpClient := &http.Client{}

	if os.Getenv("HERMIT_INSECURE") != "" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient.Transport = tr
	}

	ec, err := client.New(srv, httpClient)
	if err != nil {
		exitWithError(err)
	}

	cmds := map[string]subCommand{
		"run":     newRunCommand(ec),
		"task":    newTaskCommand(ec),
		"ls":      newListCommand(ec),
		"logs":    newLogsCommand(ec),
		"version": newVersionCommand(ec),
	}

	if len(os.Args) < 2 || os.Args[1] == "-help" || os.Args[1] == "-h" {
		usage(cmds)
		os.Exit(0)
	}

	cmd, ok := cmds[os.Args[1]]
	if !ok {
		usage(cmds)
		os.Exit(1)
	}

	cmd.Parse(os.Args[2:])
	cmd.Run()
}

type subCommand interface {
	Parse(args []string)
	Run()
}

func newFlagSet(name string, synopsis string, description string) *flag.FlagSet {
	w := os.Stdout

	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintln(w, "Usage:", synopsis)
		fmt.Fprintln(w)
		fmt.Fprintln(w, description)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options:")
		fmt.Fprintln(w)
		flags.PrintDefaults()
	}

	return flags
}

type runCommand struct {
	CPU    float64
	Memory float64
	Image  string

	flags  *flag.FlagSet
	client *client.Client
}

func newRunCommand(c *client.Client) *runCommand {
	return &runCommand{
		flags:  newFlagSet("run", "hermit run [OPTION]... COMMAND", "Run an eremetic task."),
		client: c,
	}
}

func (cmd *runCommand) Parse(args []string) {
	cmd.flags.Float64Var(&cmd.CPU, "cpu", 0.1, "CPU shares to give to the task")
	cmd.flags.Float64Var(&cmd.Memory, "mem", 128, "Memory in MB to give to the task")
	cmd.flags.StringVar(&cmd.Image, "image", "busybox", "Image to use")
	cmd.flags.Parse(args)
}

func (cmd *runCommand) Run() {
	args := cmd.flags.Args()
	if len(args) == 0 {
		cmd.flags.Usage()
		os.Exit(1)
	}

	cmdStr := strings.Join(args, " ")

	r := eremetic.Request{
		Command:     cmdStr,
		DockerImage: cmd.Image,
		TaskCPUs:    cmd.CPU,
		TaskMem:     cmd.Memory,
	}

	if err := cmd.client.AddTask(r); err != nil {
		exitWithError(err)
	}
}

type taskCommand struct {
	flags  *flag.FlagSet
	client *client.Client
}

func newTaskCommand(c *client.Client) *taskCommand {
	return &taskCommand{
		flags:  newFlagSet("task", "hermit task [OPTION]... TASK", "Fetch information about a specific task."),
		client: c,
	}
}

func (cmd *taskCommand) Parse(args []string) {
	cmd.flags.Parse(args)
}

func (cmd *taskCommand) Run() {
	taskID := cmd.flags.Arg(0)
	if taskID == "" {
		cmd.flags.Usage()
		os.Exit(1)
	}

	task, err := cmd.client.Task(taskID)
	if err != nil {
		exitWithError(err)
	}

	fmt.Println("ID:", task.ID)
	fmt.Println("Name:", task.Name)
	fmt.Println("Command:", task.Command)
	fmt.Println("Image:", task.Image)
	fmt.Println("CPU:", task.TaskCPUs)
	fmt.Println("Memory:", task.TaskMem)
	fmt.Println("State:", currentStatus(task.Status))
	fmt.Println("Last updated:", lastUpdated(task.LastUpdated()))
}

type listCommand struct {
	Filter   string
	NumTasks int
	Quiet    bool

	flags  *flag.FlagSet
	client *client.Client
}

func newListCommand(c *client.Client) *listCommand {
	return &listCommand{
		flags:  newFlagSet("ls", "hermit ls [OPTION]...", "List active eremetic tasks."),
		client: c,
	}
}

func (cmd *listCommand) Parse(args []string) {
	filters := []string{
		"min_age=<duration>",
		"max_age=<duration>",
		"status=<string>",
		"user=<string>",
	}
	filterDesc := "\n\t\t" + strings.Join(filters, "\n\t\t")

	cmd.flags.StringVar(&cmd.Filter, "filter", "", `Filter output based on conditions (example: "min_age=5m,status=running"):`+filterDesc)
	cmd.flags.IntVar(&cmd.NumTasks, "n", -1, "Show n last scheduled tasks")
	cmd.flags.BoolVar(&cmd.Quiet, "q", false, "Only display task IDs")
	cmd.flags.Parse(args)
}

func minAgeFilter(d time.Duration) func(t eremetic.Task) bool {
	return func(t eremetic.Task) bool {
		since := time.Now().Add(-d)
		return t.LastUpdated().Before(since)
	}
}

func maxAgeFilter(d time.Duration) func(t eremetic.Task) bool {
	return func(t eremetic.Task) bool {
		since := time.Now().Add(-d)
		return t.LastUpdated().After(since)
	}
}

func statusFilter(s string) func(t eremetic.Task) bool {
	return func(t eremetic.Task) bool {
		return currentStatus(t.Status) == s
	}
}

func userFilter(user string) func(t eremetic.Task) bool {
	return func(t eremetic.Task) bool {
		return t.User == user
	}
}

func filterTasks(tasks []eremetic.Task, f func(t eremetic.Task) bool) []eremetic.Task {
	var res []eremetic.Task
	for _, t := range tasks {
		if f(t) {
			res = append(res, t)
		}
	}
	return res
}

func (cmd *listCommand) Run() {
	tasks, err := cmd.client.Tasks()
	if err != nil {
		exitWithError(err)
	}

	if cmd.Filter != "" {
		kvs := strings.Split(cmd.Filter, ",")
		for _, kv := range kvs {
			kvp := strings.Split(kv, "=")

			if len(kv) < 2 {
				exitWithError(errors.New("invalid filter"))
			}

			switch kvp[0] {
			case "min_age":
				d, err := time.ParseDuration(kvp[1])
				if err != nil {
					exitWithError(err)
				}

				tasks = filterTasks(tasks, minAgeFilter(d))
			case "max_age":
				d, err := time.ParseDuration(kvp[1])
				if err != nil {
					exitWithError(err)
				}

				tasks = filterTasks(tasks, maxAgeFilter(d))
			case "status":
				tasks = filterTasks(tasks, statusFilter(kvp[1]))
			case "user":
				tasks = filterTasks(tasks, userFilter(kvp[1]))
			}
		}
	}

	if cmd.NumTasks > 0 && cmd.NumTasks < len(tasks) {
		tasks = tasks[:cmd.NumTasks]
	}

	sort.Sort(sort.Reverse(ByLastUpdated(tasks)))

	printTasks(tasks, cmd.Quiet)
}

func printTasks(tasks []eremetic.Task, quiet bool) {
	if quiet {
		for _, t := range tasks {
			fmt.Println(t.ID)
		}
		return
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)

	headers := []string{
		"TASK ID", "NAME", "STATE", "LAST UPDATED", "USER", "IMAGE", "COMMAND",
	}

	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for _, tt := range tasks {
		fields := []string{
			tt.ID, tt.Name, currentStatus(tt.Status), lastUpdated(tt.LastUpdated()), tt.User, tt.Image, fmt.Sprintf("%q", tt.Command),
		}

		fmt.Fprintln(w, strings.Join(fields, "\t"))
	}

	w.Flush()
}

type logsCommand struct {
	File string

	flags  *flag.FlagSet
	client *client.Client
}

func newLogsCommand(c *client.Client) *logsCommand {
	return &logsCommand{
		flags:  newFlagSet("logs", "hermit logs [OPTION]... COMMAND", "Fetch the logs of a task."),
		client: c,
	}
}

func (cmd *logsCommand) Parse(args []string) {
	cmd.flags.StringVar(&cmd.File, "file", "stdout", "File in the mesos sandbox")
	cmd.flags.Parse(args)
}

func (cmd *logsCommand) Run() {
	taskID := cmd.flags.Arg(0)
	if taskID == "" {
		cmd.flags.Usage()
		os.Exit(1)
	}

	b, err := cmd.client.Sandbox(taskID, cmd.File)
	if err != nil {
		exitWithError(err)
	}

	fmt.Printf("%s", b)
}

type versionCommand struct {
	flags  *flag.FlagSet
	client *client.Client
}

func newVersionCommand(c *client.Client) *versionCommand {
	return &versionCommand{
		flags:  newFlagSet("run", "hermit version", "Show the Eremetic version information."),
		client: c,
	}
}

func (cmd *versionCommand) Parse(args []string) {
	cmd.flags.Parse(args)
}

func (cmd *versionCommand) Run() {
	b, err := cmd.client.Version()
	if err != nil {
		exitWithError(err)
	}

	fmt.Println("Client:")
	fmt.Printf(" Version: %s\n", "0.0.1")
	fmt.Println()
	fmt.Println("Server:")
	fmt.Printf(" Version: %s", b)
}

type ByLastUpdated []eremetic.Task

func (t ByLastUpdated) Len() int           { return len(t) }
func (t ByLastUpdated) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t ByLastUpdated) Less(i, j int) bool { return t[i].LastUpdated().Before(t[j].LastUpdated()) }

func currentStatus(statuses []eremetic.Status) string {
	if len(statuses) == 0 {
		return ""
	}

	switch statuses[len(statuses)-1].Status {
	case eremetic.TaskState_TASK_STAGING:
		return "staging"
	case eremetic.TaskState_TASK_STARTING:
		return "starting"
	case eremetic.TaskState_TASK_RUNNING:
		return "running"
	case eremetic.TaskState_TASK_FINISHED:
		return "finished"
	case eremetic.TaskState_TASK_FAILED:
		return "failed"
	case eremetic.TaskState_TASK_KILLED:
		return "killed"
	case eremetic.TaskState_TASK_LOST:
		return "lost"
	case eremetic.TaskState_TASK_ERROR:
		return "error"
	case eremetic.TaskState_TASK_QUEUED:
		return "queued"
	}

	return "unknown"
}

func lastUpdated(t time.Time) string {
	d := time.Since(t)
	h := int64(math.Floor(d.Hours()))
	m := int64(math.Floor(d.Minutes()))
	s := int64(math.Floor(d.Seconds()))

	if h > 0 {
		return fmt.Sprintf("About %d hour(s) ago", h)
	}
	if m > 0 {
		return fmt.Sprintf("%d minute(s) ago", m)
	}
	if s > 0 {
		return fmt.Sprintf("%d second(s) ago", s)
	}

	return ""
}

func exitWithError(err error) {
	fmt.Println("Error:", err)
	os.Exit(1)
}

func usage(cmds map[string]subCommand) {
	fmt.Print("A CLI tool for communicating with Eremetic.\n\n")
	var commands []string
	for cmd := range cmds {
		commands = append(commands, cmd)
	}
	strings.Join(commands, ", ")
	fmt.Printf("Available sub-commands: %s\n\n", commands)
	fmt.Println("Use hermit <sub-command> -help for more information.")
}
