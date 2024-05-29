# unew

Append lines from stdin to a file, but only if they don't already appear in the file.
Outputs new lines to `stdout` too, making it a bit like a `tee -a` that removes duplicates.

## Usage
```
Usage of unew:
  -a    append output; do not sort
  -q    quiet mode (no output at all on terminal)
  -t    trim leading and trailing whitespace before comparison
  -v    print version information and exit
```

## Speed Comparison
```
## time cat chaos-subs.txt | unew -q subs1.txt
real    0m26.252s
user    0m28.826s
sys     0m9.321s

## time cat chaos-subs.txt | anew -q subs2.txt
real    1m2.659s
user    0m37.907s
sys     0m36.362s

## time cat chaos-subs.txt | sort -u >> subs3.txt
real    1m26.432s
user    1m11.493s
sys     0m3.562s
```

## Usage Example

Here, a file called `things.txt` contains a list of numbers. `newthings.txt` contains a second
list of numbers, some of which appear in `things.txt` and some of which do not. `unew` is used
to append the latter to `things.txt`.


```
▶ cat things.txt
Zero
One
Two

▶ cat newthings.txt
One
Two
Three
Four

▶ cat newthings.txt | unew -p things.txt
Three
Four

▶ cat things.txt
Zero
One
Two
Three
Four

```

Note that the new lines added to `things.txt` are also sent to `stdout`, this allows for them to
be redirected to another file:

```
▶ cat newthings.txt | unew things.txt > added-lines.txt
▶ cat added-lines.txt
Three
Four
```

## Flags

- To view the output in stdout, but not append to the file, use the dry-run option `-d`.
- To append to the file, but not print anything to stdout, use quiet mode `-q`.

## Install

You can either install using go:

```
go install -v github.com/rix4uni/unew@latest
```

Or
```
git clone https://github.com/rix4uni/unew.git && cd unew && go build unew.go && mv unew /usr/bin/
```

Or download a [binary release](https://github.com/rix4uni/unew/releases) for your platform.
