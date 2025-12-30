# File Cleaner - Advanced Text Processing Tool

A powerful Go-based command-line utility for cleaning and processing text files. Remove invisible characters, strip formatting, normalize line endings, and ensure your files are clean and consistent across platforms.

## Features

✨ **Character Cleaning**
- Remove zero-width and invisible Unicode characters
- Strip non-ASCII characters
- Remove control characters (preserving newlines and tabs)
- Remove Byte Order Mark (BOM)
- Normalize whitespace

🎨 **Format Stripping**
- **Markdown**: Automatically detect and strip Markdown formatting
- **HTML**: Remove HTML tags and decode entities

🔄 **Line Ending Conversion**
- Auto-detect current OS
- Convert to Windows (CRLF), Unix/Linux (LF), or Classic Mac (CR)
- Cross-platform compatibility

📊 **Detailed Reporting**
- Statistics on removed characters
- Line-by-line processing information
- Character breakdown with Unicode codes

## Installation

### Prerequisites
- Go 1.16 or higher

### Build from Source

```bash
# Clone or download the clean.go file
git clone <repository-url>
cd <repository-directory>

# Build the executable
go build -o cleanfile clean.go

# Optional: Move to PATH for system-wide access
sudo mv cleanfile /usr/local/bin/

Command-Line Options
Required Options



Option
Description



-input <file>
Input file path (required)


Optional Options



Option
Default
Description



-output <file>
input_cleaned.ext
Output file path


-ascii
true
Remove non-ASCII characters


-control
true
Remove control characters (except newlines/tabs)


-zerowidth
true
Remove zero-width characters


-bom
true
Remove Byte Order Mark (BOM)


-normalize
false
Normalize whitespace


-preserve-newlines
true
Preserve newlines when normalizing


-backup
true
Create backup of original file


-verbose
false
Show detailed processing information


-details
false
Show detailed character breakdown


-os <type>
auto
Target OS for line endings (windows, unix, mac, auto)


-strip <format>
none
Strip formatting (markdown or html)


Usage Examples
Basic Usage
# Clean a file with default settings
./cleanfile -input document.txt

# Specify custom output file
./cleanfile -input document.txt -output clean_document.txt

# Disable backup creation
./cleanfile -input document.txt -backup=false

Character Cleaning
# Remove only zero-width characters
./cleanfile -input file.txt -ascii=false -control=false

# Keep non-ASCII characters (useful for international text)
./cleanfile -input unicode.txt -ascii=false

# Normalize whitespace while preserving newlines
./cleanfile -input messy.txt -normalize -preserve-newlines

Line Ending Conversion
# Convert to Windows line endings (CRLF)
./cleanfile -input unix_file.txt -os windows

# Convert to Unix line endings (LF)
./cleanfile -input windows_file.txt -os unix

# Auto-detect current OS (default)
./cleanfile -input file.txt -os auto

Markdown Processing
# Strip Markdown formatting
./cleanfile -input README.md -strip markdown

# Strip Markdown with verbose output
./cleanfile -input document.md -strip markdown -verbose

# Strip Markdown and convert to Windows line endings
./cleanfile -input article.md -strip markdown -os windows

# Keep non-ASCII characters when stripping Markdown
./cleanfile -input international.md -strip markdown -ascii=false

Input (document.md):
# Hello World

This is **bold** and *italic* text.

- List item 1
- List item 2

[Link](https://example.com)

`inline code`

Output (document_cleaned.md):
Hello World

This is bold and italic text.

List item 1
List item 2

Link

inline code

HTML Processing
# Strip HTML tags and decode entities
./cleanfile -input page.html -strip html

# Strip HTML with detailed output
./cleanfile -input index.html -strip html -details

# Strip HTML and keep non-ASCII characters
./cleanfile -input webpage.html -strip html -ascii=false

Input (page.html):
<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
<h1>Hello &amp;amp; Welcome</h1>
<p>This is a &amp;lt;test&amp;gt; with &amp;nbsp; entities.</p>
<p>Price: &amp;pound;99.99 &amp;euro;89.99</p>
<script>alert('removed');</script>
</body>
</html>

Output (page_cleaned.html):
Test Page

Hello &amp; Welcome
This is a <test> with   entities.
Price: £99.99 €89.99

Verbose and Detailed Output
# Show processing details
./cleanfile -input file.txt -verbose

# Show character-by-character breakdown
./cleanfile -input file.txt -details

# Combine verbose and details
./cleanfile -input file.txt -verbose -details

Example Output:
======================================================================
FILE CLEANING REPORT
======================================================================

Files:
   Input:  document.txt
   Output: document_cleaned.txt

Configuration:
   Target OS:              Unix/Linux/macOS
   Line ending format:     LF (\n)

Processing Statistics:
   Lines processed:        150
   Lines with issues:      12
   Total characters:       5432

Character Removal Summary:
   Total removed:        23 characters
   Zero-width chars:     15
   Control chars:        5
   Non-ASCII chars:      3

   Removal rate: 0.42% of total characters

Detailed Character Breakdown:
----------------------------------------------------------------------
   U+200B  Zero Width Space                          10 occurrence(s)
   U+200C  Zero Width Non-Joiner                      5 occurrence(s)
   U+0000  NULL character                             3 occurrence(s)
----------------------------------------------------------------------

======================================================================
File cleaned successfully!
======================================================================

Common Use Cases
1. Clean Code Files
# Remove invisible characters from source code
./cleanfile -input main.go -verbose

# Ensure Unix line endings for Git repositories
./cleanfile -input script.sh -os unix

2. Process Documentation
# Convert Markdown to plain text
./cleanfile -input README.md -strip markdown -output README.txt

# Clean HTML documentation
./cleanfile -input docs.html -strip html -output docs.txt

3. Prepare Files for Cross-Platform Use
# Convert Unix file for Windows
./cleanfile -input unix_script.sh -os windows -output windows_script.bat

# Convert Windows file for Unix
./cleanfile -input windows_file.txt -os unix -output unix_file.txt

4. Debug Invisible Characters
# Find and remove problematic characters
./cleanfile -input problematic.txt -verbose -details

# Check what would be removed (with backup)
./cleanfile -input file.txt -verbose -details
# Then check the .bak file to compare

5. Batch Processing
# Process multiple files
for file in *.md; do
    ./cleanfile -input "$file" -strip markdown
done

# Convert all text files to Unix format
for file in *.txt; do
    ./cleanfile -input "$file" -os unix
done

Format Detection
The tool automatically detects file formats before stripping:
Markdown Detection

Headers (#, ##, ###)
Lists (-, *, +, 1.)
Code blocks (```)
Links ([text](url))
Bold/Italic (**bold**, *italic*)

HTML Detection

DOCTYPE declarations
HTML tags (<tag>)
HTML entities (&amp;amp;, &amp;#123;)

Note: The tool will refuse to strip if the detected format doesn't match the requested format, preventing accidental data loss.
Supported HTML Entities
The tool decodes common HTML entities including:



Entity
Character
Description



&amp;amp;
&amp;
Ampersand


&amp;lt;
<
Less than


&amp;gt;
>
Greater than


&amp;quot;
"
Quote


&amp;nbsp;
(space)
Non-breaking space


&amp;copy;
©
Copyright


&amp;euro;
€
Euro


&amp;pound;
£
Pound


&amp;#65;
A
Numeric (decimal)


&amp;#x41;
A
Numeric (hexadecimal)


...and many more!
Zero-Width Characters Removed
The tool removes these invisible Unicode characters:

U+200B - Zero Width Space
U+200C - Zero Width Non-Joiner
U+200D - Zero Width Joiner
U+200E - Left-to-Right Mark
U+200F - Right-to-Left Mark
U+FEFF - BOM/Zero Width No-Break Space
U+202A-202E - Directional formatting
U+2060-2064 - Invisible operators
U+206A-206F - Shape controls

Error Handling
The tool provides clear error messages:
# Missing input file
$ ./cleanfile
Error: Input file is required

# File doesn't exist
$ ./cleanfile -input nonexistent.txt
Error: Input file 'nonexistent.txt' does not exist

# Wrong format detected
$ ./cleanfile -input code.go -strip markdown
Error: file does not appear to be Markdown (detected: unknown)

# Invalid OS option
$ ./cleanfile -input file.txt -os invalid
Error: Invalid target OS 'invalid'. Valid options: windows, unix, mac, auto

Backup Files
By default, the tool creates a backup with .bak extension:
# Original file is backed up
./cleanfile -input important.txt
# Creates: important.txt.bak

# Disable backup
./cleanfile -input file.txt -backup=false

Performance Tips

Large Files: The tool processes files line-by-line for memory efficiency
Batch Processing: Use shell loops for multiple files
Regex Compilation: Patterns are compiled once for optimal performance
Buffer Writing: Output is buffered for faster I/O

Troubleshooting
Issue: "File does not appear to be Markdown/HTML"
Solution: The detection threshold requires at least ~14% of lines to have format indicators. For files with minimal formatting, you may need to manually verify the format or adjust your expectations.
Issue: Characters still appearing after cleaning
Solution: Use -details flag to see exactly what's being removed. Some characters might be intentional or require different flags.
Issue: Line endings not converting
Solution: Ensure you're using the correct -os flag. Use -verbose to see conversion statistics.
Quick Reference
# Most common commands
./cleanfile -input file.txt                           # Basic cleaning
./cleanfile -input file.md -strip markdown            # Strip Markdown
./cleanfile -input file.html -strip html              # Strip HTML
./cleanfile -input file.txt -os windows               # Convert line endings
./cleanfile -input file.txt -verbose -details         # Debug mode
./cleanfile -input file.txt -ascii=false              # Keep Unicode

License
MIT License (or your preferred license)
Contributing
Contributions are welcome! Please submit pull requests or open issues for bugs and feature requests.
Author
[Your Name/Organization]
Version
1.0.0

Need Help? Run ./cleanfile -h to see all available options.
```
