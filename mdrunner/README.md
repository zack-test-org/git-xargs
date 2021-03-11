# mdrunner

## Build it

```
make
```

## Run it

Choose `run` every time.

```
bin/mdrunner examples/working.md
```

## Run and fix on the fly

```
./mdrunner ./broken_example.md
```

1. `run` the first and second steps, watch it break.
1. On the second run, watch it pick up where it left off.
1. Play around with `view`, `skip`, and `edit`.

## Demo

```
$ bin/mdrunner examples/working.md


############## Step 1: Get the date and export it as CURR_DATE


# Get the date and export it as CURR_DATE

date=$(date +%Y-%m-%d)
echo "The current date (YY-MM-DD) is $date"
echo "Exporting this date as CURR_DATE"
export CURR_DATE="${date}"

##############################


? # Step 1: Get the date and export it as CURR_DATE run
# Running Step 1: Get the date and export it as CURR_DATE

The current date (YY-MM-DD) is 2021-02-20
Exporting this date as CURR_DATE



############## Step 2: Show the date once more and exit cleanly


# Show the date once more and exit cleanly

echo "The current date is ${CURR_DATE}."

##############################


? # Step 2: Show the date once more and exit cleanly run
# Running Step 2: Show the date once more and exit cleanly

The current date is 2021-02-20.
```
