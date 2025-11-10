## unew

A high-performance command-line utility for processing and managing unique lines from input streams. Combining the functionality of `sort`, `uniq`, and `tee`, `unew` efficiently filters duplicates while offering advanced features like file splitting, shuffling, and case-insensitive processing.

## 🚀 Performance

Benchmarked against similar tools, `unew` demonstrates significant speed advantages:

```yaml
# Processing a large file with 1M+ lines
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

## 📦 Installation

### Quick Install (Go)
```yaml
go install github.com/rix4uni/unew@latest
```

### Prebuilt Binaries
Download the latest release for your platform:
```yaml
# Linux
wget https://github.com/rix4uni/unew/releases/download/v0.0.7/unew-linux-amd64-0.0.7.tgz
tar -xvzf unew-linux-amd64-0.0.7.tgz
sudo mv unew /usr/local/bin/

# Or manually download from:
# https://github.com/rix4uni/unew/releases
```

### From Source
```yaml
git clone https://github.com/rix4uni/unew.git
cd unew
go build -o unew .
sudo mv unew /usr/local/bin/
```

## ⚡ Quick Start

### Basic Deduplication
```yaml
# Remove duplicates while preserving input order
cat input.txt | unew

# Save to file and suppress terminal output
cat input.txt | unew -q output.txt

# Append new unique lines to existing file
cat new_data.txt | unew -a -q existing.txt
```

### Advanced Processing
```yaml
# Case-insensitive deduplication with whitespace trimming
cat data.txt | unew -i -t -el

# Shuffle output lines randomly
cat list.txt | unew -shuf shuffled.txt
```

## 🔧 Command Reference

| Flag | Description | Example |
|------|-------------|---------|
| `-a` | Append mode (disables deduplication, only with `-q`) | `unew -a -q file.txt` |
| `-divide N` | Split into N equal files (N ≥ 2) | `unew -divide 3 prefix_` |
| `-ef` | Prevent empty file creation | `unew -ef output.txt` |
| `-el` | Skip empty lines from input | `unew -el` |
| `-i` | Case-insensitive comparison | `unew -i` |
| `-q` | Quiet mode (suppress stdout) | `unew -q file.txt` |
| `-shuf` | Randomly shuffle output | `unew -shuf` |
| `-size SIZE` | Split by file size | `unew -size 1GB data_` |
| `-split N` | Split every N lines | `unew -split 1000 chunks_` |
| `-t` | Trim whitespace from lines | `unew -t` |
| `-version` | Show version information | `unew -version` |

## 📁 File Splitting Modes

### Line-based Splitting
```yaml
# Create files with 1000 lines each
cat large_list.txt | unew -split 1000 part_

# Results: part1.txt, part2.txt, part3.txt...
```

### Equal Division
```yaml
# Distribute lines evenly across 4 files
cat data.txt | unew -divide 4 segment_

# Each file gets roughly 25% of the lines
```

### Size-based Splitting
```yaml
# Split when files reach specified size
cat big_file.txt | unew -size 500MB archive_

# Supported units: B, KB, MB, GB
unew -size 1024KB   # 1MB chunks
unew -size 2GB      # 2GB chunks
```

## 🎯 Use Cases

### Security & Reconnaissance
```yaml
# Process subdomain lists with deduplication
subfinder -d example.com | unew -i -t -el all_subs.txt

# Merge and deduplicate multiple wordlists
cat wordlist1.txt wordlist2.txt | unew -i combined.txt
```

### Data Processing
```yaml
# Clean and deduplicate CSV data
cut -d, -f1 data.csv | unew -t -el unique_values.txt

# Shuffle training data for machine learning
cat training_data.txt | unew -shuf shuffled_data.txt
```

### Log Analysis
```yaml
# Extract unique IP addresses from logs
cat access.log | awk '{print $1}' | unew unique_ips.txt

# Process case-insensitive error messages
cat app.log | grep ERROR | unew -i -t error_types.txt
```

## 🔍 Examples

### Input Processing
```yaml
# Sample input file
▶ cat domains.txt
example.com
 Example.COM   
admin.example.com

admin.example.com
TEST.EXAMPLE.COM

# Basic deduplication
▶ cat domains.txt | unew
example.com
 Example.COM   
admin.example.com
TEST.EXAMPLE.COM

# With trimming and case sensitivity
▶ cat domains.txt | unew -t -i
example.com
admin.example.com

# Remove empty lines and trim
▶ cat domains.txt | unew -t -el
example.com
Example.COM
admin.example.com
TEST.EXAMPLE.COM
```

### File Management
```yaml
# Append new lines without duplicates
cat new_domains.txt | unew -a -q existing_domains.txt

# Split large password list
cat rockyou.txt | unew -split 50000 passwords_

# Divide massive dataset
cat huge_list.txt | unew -divide 10 chunk_
```

## 💡 Pro Tips

1. **Combine Flags**: Use `-t -i -el` for comprehensive data cleaning
2. **Memory Efficient**: `unew` handles large files without loading everything into memory
3. **Pipe Friendly**: Perfect for chaining with other Unix tools
4. **File Safety**: Use `-ef` to avoid creating empty files in automated scripts

## 🐛 Troubleshooting

**Common Issues:**
- `-divide` requires N ≥ 2
- `-a` mode only works with `-q`
- Filename prefix required for splitting operations
- Size units must be uppercase (GB, MB, KB, B)

**Example Error:**
```yaml
▶ cat data.txt | unew -divide 1 output_
Error: -divide flag must be >= 2

▶ cat data.txt | unew -a -t
Error: -q flag is the only flag allowed with -a flag
```