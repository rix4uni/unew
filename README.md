## unew

A tool combined of 2 commands features in 1 `sort` and `tee` for adding new lines to files, skipping duplicates

## Installation
```
go install github.com/rix4uni/unew@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/unew/releases/download/v0.0.6/unew-linux-amd64-0.0.6.tgz
tar -xvzf unew-linux-amd64-0.0.6.tgz
rm -rf unew-linux-amd64-0.0.6.tgz
mv unew ~/go/bin/unew
```
Or download [binary release](https://github.com/rix4uni/unew/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/unew.git
cd unew; go install
```

## Usage
```
Usage of unew:
  -a    append output; do not sort
  -ef
        do not create empty files
  -el
        remove empty lines from input
  -i    ignore case during comparison
  -q    quiet mode (no output at all on terminal)
  -shuf
        shuffle the output lines randomly
  -split int
        split the output into files with a specified number of lines per file
  -t    trim leading and trailing whitespace before comparison
  -version
        print version information and exit
```

## Speed Comparison
```
▶ time cat chaos-subs.txt | unew -q subs1.txt
real    0m26.252s
user    0m28.826s
sys     0m9.321s

▶ time cat chaos-subs.txt | anew -q subs2.txt
real    1m2.659s
user    0m37.907s
sys     0m36.362s

▶ time cat chaos-subs.txt | sort -u > subs3.txt
real    1m26.432s
user    1m11.493s
sys     0m3.562s
```

## Usage Examples
```
# input
▶ cat things.txt
rix4uni.com

admin.rix4uni.com
admin.rix4uni.com
JENKINS.rix4uni.com

# output e.g. 1
▶ cat things.txt | unew -a
rix4uni.com

admin.rix4uni.com
admin.rix4uni.com
JENKINS.rix4uni.com

# output e.g. 2
▶ cat things.txt | unew
rix4uni.com

admin.rix4uni.com
JENKINS.rix4uni.com

# output e.g. 3
▶ cat things.txt | unew -el
rix4uni.com
admin.rix4uni.com
JENKINS.rix4uni.com

# output e.g. 4
▶ cat things.txt | unew -el -i
rix4uni.com
admin.rix4uni.com
jenkins.rix4uni.com

# output e.g. 5
▶ cat things.txt | unew -split 100 newthings_split_.txt
```
