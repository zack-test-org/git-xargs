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

# Edit, re-run, or skip steps

At each step of the way, a step can be edited prior to running,
re-run, or skipped.

```bash

# Show the date once more and exit cleanly

echo "The current date is ${CURR_DATE}."
```
