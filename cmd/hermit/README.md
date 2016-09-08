# hermit

A tool for running and listing Eremetic tasks from the command-line.

## Examples:

Run an Eremetic task.

    hermit run -cpu 0.2 -mem 32 -image busybox echo hello

List active Eremetic tasks.
    
    hermit ls

Fetch information about a specific task.
    
    hermit task eremetic-task-id-abc123

Fetch the logs of a task.

    hermit logs -file stderr eremetic-task-id-abc123

## Configuration

You can configure hermit using these environment variables:

- `EREMETIC_URL`: URL of Eremetic server to connect to.
- `HERMIT_INSECURE`: Allow establishing insecure connections.
