# mdrunner - a tool to iterate through code blocks in a Markdown file

`mdrunner` loads in code snippets from a Markdown file and runs them sequentially, preserving directory and environment state between steps. If an error occurs, when you re-run `mdrunner` it will pick back up where it left off with the working directory and environment state preserved.

At any step, you are given the option to run the current step, skip it, or edit it prior to running.

## Installation

There is a simple make file to build the binary in the `bin/` directory.

```
make
```

## A Working example

```
$ bin/mdrunner examples/working.md


############## Step 1: Get the date and export it as CURR_DATE


# Get the date and export it as CURR_DATE
cd /tmp
date=$(date +%Y-%m-%d)
echo "The current date (YY-MM-DD) is $date"
echo "Exporting this date as CURR_DATE"
export CURR_DATE="${date}"

##############################


? # Step 1: Get the date and export it as CURR_DATE  [Use arrows to move, type to filter]
> run
  skip
  edit
  quit
  ```

  Choose `run`.

  ```
  ? # Step 1: Get the date and export it as CURR_DATE run
# Running Step 1: Get the date and export it as CURR_DATE

The current date (YY-MM-DD) is 2021-03-11
Exporting this date as CURR_DATE



############## Step 2: Show the date once more and exit cleanly


# Show the date once more and exit cleanly

echo "The current date is ${CURR_DATE}."

##############################


? # Step 2: Show the date once more and exit cleanly  [Use arrows to move, type to filter]
> run
  skip
  edit
  quit
  ```

Again, choose `run`.

```
# Running Step 2: Show the date once more and exit cleanly

The current date is 2021-03-11.
```

# A "broken" example

```
$ bin/mdrunner examples/broken.md


############## Step 1: Get the date and export it as CURR_DATE


# Get the date and export it as CURR_DATE

cd /tmp
date=$(date +%Y-%m-%d)
echo "The current date (YY-MM-DD) is $date"
echo "Exporting this date as CURR_DATE"
export CURR_DATE="${date}"

##############################


? # Step 1: Get the date and export it as CURR_DATE  [Use arrows to move, type to filter]
> run
  skip
  edit
  quit
  ```

  Choose `run`.

```
# Running Step 1: Get the date and export it as CURR_DATE

The current date (YY-MM-DD) is 2021-03-11
Exporting this date as CURR_DATE



############## Step 2: Show the current date, but exit badly to demonstrate failure


# Show the current date, but exit badly to demonstrate failure

echo "From the previous step, the date is ${CURR_DATE}"
echo "We're going to break the process by exiting with a non-zero code"
exit 1

##############################


? # Step 2: Show the current date, but exit badly to demonstrate failure  [Use arrows to move, type to filter]
> run
  skip
  edit
  quit
```

Again, choose `run`.

```
# Running Step 2: Show the current date, but exit badly to demonstrate failure

From the previous step, the date is 2021-03-11
We're going to break the process by exiting with a non-zero code

Saved state file /tmp/mdrunner-b8397b3.yaml

FATA[0078] exit status 1
```

Run `mdrunner` again. Note that it picks up where it left off. Use `edit` to fix the error (remove lines 5 and 6, the bad comment and the `exit 1`), then `run` the step again.

```
$ bin/mdrunner examples/broken.md
Loading state from /tmp/mdrunner-b8397b3.yaml

# Skipping Step 1: Get the date and export it as CURR_DATE


############## Step 2: Show the current date, but exit badly to demonstrate failure


# Show the current date, but exit badly to demonstrate failure

echo "From the previous step, the date is ${CURR_DATE}"
echo "We're going to break the process by exiting with a non-zero code"
exit 1

##############################


? # Step 2: Show the current date, but exit badly to demonstrate failure  [Use arrows to move, type to filter]
  run
  skip
> edit
  quit
```

Edit the file to remove the two lines of code.

```
############## Step 2: Show the current date, but exit badly to demonstrate failure


# Show the current date, but exit badly to demonstrate failure

echo "From the previous step, the date is ${CURR_DATE}"


##############################


? # Step 2: Show the current date, but exit badly to demonstrate failure run
# Running Step 2: Show the current date, but exit badly to demonstrate failure

From the previous step, the date is 2021-03-11



############## Step 3: Finally, show the date once more and exit cleanly


# Finally, show the date once more and exit cleanly

echo "The current date is ${CURR_DATE}."

##############################


? # Step 3: Finally, show the date once more and exit cleanly run
# Running Step 3: Finally, show the date once more and exit cleanly

The current date is 2021-03-11.
```

## How to create a working Markdown file

It's possible that your existing Markdown files may work without further modification, but if they do need to change, the changes are minimal.

If a code block starts with ` ```bash `, then it will be detected and run. The first comment will be given to the user as a prompt. Directory and environment variables are preserved between steps, and if a step fails, the variables will be available on re-run.

See the files in the [examples](examples/) directory.

## Automated testing

To test your documentation, you can supply the `--no-prompt` flag. This will run every step sequentially until it gets to the end. If your exit code is `0`, then your test is a success!
## Help

```
$ bin/mdrunner --help
Usage:
  mdrunner FILE [flags]

Flags:
      --cleanup            Only run the steps marked by 'mdrunner:cleanup' (not implemented yet)
  -h, --help               help for mdrunner
      --log-level string   Log level (DEBUG, INFO, WARN, ERROR, FATAL, PANIC) (can be set by MDRUNNER_LOGLEVEL) (default "WARN")
      --no-cleanup         Prevent the cleanup step marked by 'mdrunner:cleanup' from running
      --no-prompt          Iterate over scripts without pausing (useful for automated testing)
      --reset              Clear state prior to running (not implemented yet)
      --step ints          Run only the specified step
      --verify             Only run the steps marked by 'mdrunner:verify'
```



