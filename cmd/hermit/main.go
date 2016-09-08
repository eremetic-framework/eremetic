package main

import (
	"crypto/tls"
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

	cmd, ok := cmds[os.Args[1]]
	if !ok {
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
	User string

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
	cmd.flags.StringVar(&cmd.User, "user", "", "List tasks for a given user")
	cmd.flags.Parse(args)
}

func (cmd *listCommand) Run() {
	tasks, err := cmd.client.Tasks()
	if err != nil {
		exitWithError(err)
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)

	headers := []string{
		"TASK ID", "NAME", "STATE", "LAST UPDATED", "USER", "IMAGE", "COMMAND",
	}

	fmt.Fprintln(w, strings.Join(headers, "\t"))

	sort.Sort(sort.Reverse(ByLastUpdated(tasks)))

	for _, tt := range tasks {
		if cmd.User != "" && tt.User != cmd.User {
			continue
		}

		fields := []string{
			tt.ID, tt.Name, currentStatus(tt.Status), lastUpdated(tt.LastUpdated()), tt.User, tt.Image, fmt.Sprintf("%q", tt.Command),
		}

		fmt.Fprintln(w, strings.Join(fields, "\t"))
	}

	fmt.Fprintln(w)
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

	fmt.Printf("%s\n", b)
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
