# Example Markdown to be run

## Each code section is run independently

There should be no surprises, here

```bash

# Get the date and export it as CURR_DATE

cd /tmp
date=$(date +%Y-%m-%d)
echo "The current date (YY-MM-DD) is $date"
echo "Exporting this date as CURR_DATE"
export CURR_DATE="${date}"
```

# We retain state from section to section

Any `export FOO="bar" line will be retained from step to stop`

```bash

# Show the current date, but exit badly to demonstrate failure

echo "From the previous step, the date is ${CURR_DATE}"
echo "We're going to break the process by exiting with a non-zero code"
exit 1
```

# Edit, re-run, or skip steps

At each step of the way, a step can be edited prior to running,
re-run, or skipped.

```bash

# Finally, show the date once more and exit cleanly

echo "The current date is ${CURR_DATE}."
```
