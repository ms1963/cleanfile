package main

import (
        "bufio"
        "flag"
        "fmt"
        "io"
        "os"
        "path/filepath"
        "regexp"
        "runtime"
        "strings"
        "unicode"
)

// CleaningOptions defines what types of characters to remove
type CleaningOptions struct {
        RemoveNonASCII      bool
        RemoveControlChars  bool
        RemoveZeroWidth     bool
        RemoveBOM           bool
        NormalizeWhitespace bool
        PreserveNewlines    bool
        TargetOS            string
        StripFormat         string
}

// CleaningStats holds statistics about the cleaning process
type CleaningStats struct {
        TotalChars           int
        RemovedChars         int
        NonASCIIRemoved      int
        ControlCharsRemoved  int
        ZeroWidthRemoved     int
        LinesProcessed       int
        LinesWithIssues      int
        LineEndingsConverted int
        RemovedCharDetails   map[rune]int
        MarkdownStripped     bool
        HTMLStripped         bool
        HTMLEntitiesDecoded  int
        FormatDetected       string
}

// Common zero-width and invisible Unicode characters
var zeroWidthChars = []rune{
        '\u200B', '\u200C', '\u200D', '\u200E', '\u200F', '\uFEFF',
        '\u202A', '\u202B', '\u202C', '\u202D', '\u202E', '\u2060',
        '\u2061', '\u2062', '\u2063', '\u2064', '\u206A', '\u206B',
        '\u206C', '\u206D', '\u206E', '\u206F',
}

// HTML entity mappings
var htmlEntities = map[string]string{
        "&amp;":    "&",
        "&lt;":     "<",
        "&gt;":     ">",
        "&quot;":   "\"",
        "&apos;":   "'",
        "&nbsp;":   " ",
        "&copy;":   "\u00A9",
        "&reg;":    "\u00AE",
        "&trade;":  "\u2122",
        "&euro;":   "\u20AC",
        "&pound;":  "\u00A3",
        "&yen;":    "\u00A5",
        "&cent;":   "\u00A2",
        "&sect;":   "\u00A7",
        "&para;":   "\u00B6",
        "&middot;": "\u00B7",
        "&bull;":   "\u2022",
        "&hellip;": "\u2026",
        "&ndash;":  "\u2013",
        "&mdash;":  "\u2014",
        "&lsquo;":  "\u2018",
        "&rsquo;":  "\u2019",
        "&ldquo;":  "\u201C",
        "&rdquo;":  "\u201D",
        "&times;":  "\u00D7",
        "&divide;": "\u00F7",
        "&deg;":    "\u00B0",
        "&plusmn;": "\u00B1",
        "&frac14;": "\u00BC",
        "&frac12;": "\u00BD",
        "&frac34;": "\u00BE",
}

// Character descriptions for better output
var charDescriptions = map[rune]string{
        '\u200B': "Zero Width Space",
        '\u200C': "Zero Width Non-Joiner",
        '\u200D': "Zero Width Joiner",
        '\u200E': "Left-to-Right Mark",
        '\u200F': "Right-to-Left Mark",
        '\uFEFF': "BOM/Zero Width No-Break Space",
        '\u202A': "Left-to-Right Embedding",
        '\u202B': "Right-to-Left Embedding",
        '\u202C': "Pop Directional Formatting",
        '\u202D': "Left-to-Right Override",
        '\u202E': "Right-to-Left Override",
        '\u2060': "Word Joiner",
        '\u0000': "NULL character",
        '\u0001': "Start of Heading",
        '\u0002': "Start of Text",
        '\u0003': "End of Text",
        '\u0004': "End of Transmission",
        '\u0005': "Enquiry",
        '\u0006': "Acknowledge",
        '\u0007': "Bell",
        '\u0008': "Backspace",
        '\u000B': "Vertical Tab",
        '\u000C': "Form Feed",
        '\u000E': "Shift Out",
        '\u000F': "Shift In",
        '\r':     "Carriage Return (CR)",
        '\n':     "Line Feed (LF)",
}

func main() {
        inputFile := flag.String("input", "", "Input file path (required)")
        outputFile := flag.String("output", "", "Output file path (defaults to input_cleaned.ext)")
        removeNonASCII := flag.Bool("ascii", true, "Remove non-ASCII characters")
        removeControl := flag.Bool("control", true, "Remove control characters (except newlines/tabs)")
        removeZeroWidth := flag.Bool("zerowidth", true, "Remove zero-width characters")
        removeBOM := flag.Bool("bom", true, "Remove Byte Order Mark (BOM)")
        normalizeWS := flag.Bool("normalize", false, "Normalize whitespace")
        preserveNL := flag.Bool("preserve-newlines", true, "Preserve newlines when normalizing")
        backup := flag.Bool("backup", true, "Create backup of original file")
        verbose := flag.Bool("verbose", false, "Verbose output")
        showDetails := flag.Bool("details", false, "Show detailed list of removed characters")
        targetOS := flag.String("os", "", "Target OS for line endings (windows, unix, mac, auto). Default: auto")
        stripFormat := flag.String("strip", "", "Strip formatting: 'markdown' or 'html'")

        flag.Parse()

        if *inputFile == "" {
                fmt.Println("Error: Input file is required")
                fmt.Println("\nUsage:")
                flag.PrintDefaults()
                os.Exit(1)
        }

        if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
                fmt.Printf("Error: Input file '%s' does not exist\n", *inputFile)
                os.Exit(1)
        }

        *stripFormat = strings.ToLower(strings.TrimSpace(*stripFormat))
        if *stripFormat != "" && *stripFormat != "markdown" && *stripFormat != "html" {
                fmt.Printf("Error: Invalid strip format '%s'. Valid options: markdown, html\n", *stripFormat)
                os.Exit(1)
        }

        if *outputFile == "" {
                ext := filepath.Ext(*inputFile)
                base := strings.TrimSuffix(*inputFile, ext)
                if ext == "" {
                        *outputFile = base + "_cleaned"
                } else {
                        *outputFile = base + "_cleaned" + ext
                }
        }

        absInput, err := filepath.Abs(*inputFile)
        if err != nil {
                fmt.Printf("Error: Could not resolve input file path: %v\n", err)
                os.Exit(1)
        }
        absOutput, err := filepath.Abs(*outputFile)
        if err != nil {
                fmt.Printf("Error: Could not resolve output file path: %v\n", err)
                os.Exit(1)
        }
        if absInput == absOutput {
                fmt.Println("Error: Output file cannot be the same as input file")
                os.Exit(1)
        }

        normalizedOS := normalizeTargetOS(*targetOS)
        if normalizedOS == "" {
                fmt.Printf("Error: Invalid target OS '%s'. Valid options: windows, unix, mac, auto\n", *targetOS)
                os.Exit(1)
        }

        options := CleaningOptions{
                RemoveNonASCII:      *removeNonASCII,
                RemoveControlChars:  *removeControl,
                RemoveZeroWidth:     *removeZeroWidth,
                RemoveBOM:           *removeBOM,
                NormalizeWhitespace: *normalizeWS,
                PreserveNewlines:    *preserveNL,
                TargetOS:            normalizedOS,
                StripFormat:         *stripFormat,
        }

        if *backup {
                backupPath := *inputFile + ".bak"
                if err := copyFile(*inputFile, backupPath); err != nil {
                        fmt.Printf("Warning: Could not create backup: %v\n", err)
                } else if *verbose {
                        fmt.Printf("Backup created: %s\n", backupPath)
                }
        }

        stats, err := cleanFile(*inputFile, *outputFile, options, *verbose)
        if err != nil {
                fmt.Printf("Error: %v\n", err)
                os.Exit(1)
        }

        printResults(*inputFile, *outputFile, stats, *showDetails, normalizedOS)
}

func normalizeTargetOS(targetOS string) string {
        targetOS = strings.ToLower(strings.TrimSpace(targetOS))

        if targetOS == "" || targetOS == "auto" {
                switch runtime.GOOS {
                case "windows":
                        return "windows"
                case "darwin":
                        return "unix"
                default:
                        return "unix"
                }
        }

        switch targetOS {
        case "windows", "win", "dos":
                return "windows"
        case "unix", "linux":
                return "unix"
        case "mac", "macos", "darwin":
                return "unix"
        case "mac9", "macos9", "classic":
                return "mac9"
        default:
                return ""
        }
}

func getLineEnding(targetOS string) string {
        switch targetOS {
        case "windows":
                return "\r\n"
        case "unix":
                return "\n"
        case "mac9":
                return "\r"
        default:
                return "\n"
        }
}

func detectFileFormat(content string) string {
        if len(content) == 0 {
                return "unknown"
        }

        lines := strings.Split(content, "\n")
        markdownScore := 0
        htmlScore := 0
        totalNonEmptyLines := 0

        htmlTagPattern := regexp.MustCompile(`<[a-zA-Z][^>]*>`)
        doctypePattern := regexp.MustCompile(`(?i)<!DOCTYPE\s+html`)
        htmlEntityPattern := regexp.MustCompile(`&[a-zA-Z]+;|&#\d+;|&#x[0-9a-fA-F]+;`)
        markdownHeaderPattern := regexp.MustCompile(`^#{1,6}\s+.+`)
        markdownListPattern := regexp.MustCompile(`^[\s]*[-*+]\s+.+`)
        markdownOrderedListPattern := regexp.MustCompile(`^[\s]*\d+\.\s+.+`)
        markdownCodePattern := regexp.MustCompile("^```")
        markdownLinkPattern := regexp.MustCompile(`\[.+?\]\(.+?\)`)
        markdownBoldPattern := regexp.MustCompile(`\*\*.+?\*\*|__.+?__`)
        markdownItalicPattern := regexp.MustCompile(`\*.+?\*|_.+?_`)
        inlineCodePattern := regexp.MustCompile("`[^`]+`")

        inCodeBlock := false

        for _, line := range lines {
                trimmed := strings.TrimSpace(line)
                if trimmed == "" {
                        continue
                }

                totalNonEmptyLines++

                if markdownCodePattern.MatchString(trimmed) {
                        inCodeBlock = !inCodeBlock
                }

                if doctypePattern.MatchString(trimmed) {
                        htmlScore += 15
                }
                if htmlTagPattern.MatchString(trimmed) {
                        htmlScore += 3
                }
                if htmlEntityPattern.MatchString(trimmed) {
                        htmlScore += 2
                }

                if markdownHeaderPattern.MatchString(trimmed) {
                        markdownScore += 4
                }
                if markdownListPattern.MatchString(trimmed) {
                        markdownScore += 3
                }
                if markdownOrderedListPattern.MatchString(trimmed) {
                        markdownScore += 3
                }
                if markdownCodePattern.MatchString(trimmed) {
                        markdownScore += 4
                }
                if !inCodeBlock {
                        if markdownLinkPattern.MatchString(trimmed) {
                                markdownScore += 3
                        }
                        if markdownBoldPattern.MatchString(trimmed) {
                                markdownScore += 2
                        }
                        if markdownItalicPattern.MatchString(trimmed) {
                                markdownScore += 1
                        }
                        if inlineCodePattern.MatchString(trimmed) {
                                markdownScore += 1
                        }
                }
        }

        if totalNonEmptyLines == 0 {
                return "unknown"
        }

        threshold := totalNonEmptyLines / 7

        if htmlScore > markdownScore && htmlScore >= threshold {
                return "html"
        } else if markdownScore > htmlScore && markdownScore >= threshold {
                return "markdown"
        }

        return "unknown"
}

func stripMarkdown(text string) string {
        codeBlockPattern := regexp.MustCompile("(?s)```[a-zA-Z]*\n(.*?)```")
        text = codeBlockPattern.ReplaceAllString(text, "$1")

        inlineCodePattern := regexp.MustCompile("`([^`]+)`")
        text = inlineCodePattern.ReplaceAllString(text, "$1")

        headerPattern := regexp.MustCompile(`(?m)^#{1,6}\s+(.+)$`)
        text = headerPattern.ReplaceAllString(text, "$1")

        boldPattern := regexp.MustCompile(`(\*\*|__)(.*?)\1`)
        text = boldPattern.ReplaceAllString(text, "$2")

        italicPattern := regexp.MustCompile(`(\*|_)(.*?)\1`)
        text = italicPattern.ReplaceAllString(text, "$2")

        strikePattern := regexp.MustCompile(`~~(.*?)~~`)
        text = strikePattern.ReplaceAllString(text, "$1")

        imagePattern := regexp.MustCompile(`!\[([^\]]*)\]\([^\)]+\)`)
        text = imagePattern.ReplaceAllString(text, "$1")

        linkPattern := regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
        text = linkPattern.ReplaceAllString(text, "$1")

        refLinkPattern := regexp.MustCompile(`\[([^\]]+)\]\[[^\]]*\]`)
        text = refLinkPattern.ReplaceAllString(text, "$1")

        linkDefPattern := regexp.MustCompile(`(?m)^\[.+?\]:\s+.+$`)
        text = linkDefPattern.ReplaceAllString(text, "")

        hrPattern := regexp.MustCompile(`(?m)^[\s]*[\-\*_]{3,}[\s]*$`)
        text = hrPattern.ReplaceAllString(text, "")

        blockquotePattern := regexp.MustCompile(`(?m)^>\s*(.*)$`)
        text = blockquotePattern.ReplaceAllString(text, "$1")

        listPattern := regexp.MustCompile(`(?m)^[\s]*[-*+]\s+(.+)$`)
        text = listPattern.ReplaceAllString(text, "$1")

        orderedListPattern := regexp.MustCompile(`(?m)^[\s]*\d+\.\s+(.+)$`)
        text = orderedListPattern.ReplaceAllString(text, "$1")

        taskListPattern := regexp.MustCompile(`(?m)^[\s]*-\s*\[[xX\s]\]\s+(.+)$`)
        text = taskListPattern.ReplaceAllString(text, "$1")

        text = strings.ReplaceAll(text, "|", " ")

        htmlCommentPattern := regexp.MustCompile(`<!--.*?-->`)
        text = htmlCommentPattern.ReplaceAllString(text, "")

        multipleNewlinesPattern := regexp.MustCompile(`\n{3,}`)
        text = multipleNewlinesPattern.ReplaceAllString(text, "\n\n")

        return text
}

func stripHTML(text string) (string, int) {
        entitiesDecoded := 0

        commentPattern := regexp.MustCompile(`<!--[\s\S]*?-->`)
        text = commentPattern.ReplaceAllString(text, "")

        scriptPattern := regexp.MustCompile(`(?is)<script[^>]*>[\s\S]*?</script>`)
        text = scriptPattern.ReplaceAllString(text, "")
        stylePattern := regexp.MustCompile(`(?is)<style[^>]*>[\s\S]*?</style>`)
        text = stylePattern.ReplaceAllString(text, "")

        tagPattern := regexp.MustCompile(`<[^>]+>`)
        text = tagPattern.ReplaceAllString(text, "")

        for entity, char := range htmlEntities {
                if strings.Contains(text, entity) {
                        count := strings.Count(text, entity)
                        entitiesDecoded += count
                        text = strings.ReplaceAll(text, entity, char)
                }
        }

        numericEntityPattern := regexp.MustCompile(`&#(\d+);`)
        text = numericEntityPattern.ReplaceAllStringFunc(text, func(match string) string {
                var code int
                n, err := fmt.Sscanf(match, "&#%d;", &code)
                if err == nil && n == 1 && code > 0 && code <= 0x10FFFF {
                        entitiesDecoded++
                        return string(rune(code))
                }
                return match
        })

        hexEntityPattern := regexp.MustCompile(`(?i)&#x([0-9a-fA-F]+);`)
        text = hexEntityPattern.ReplaceAllStringFunc(text, func(match string) string {
                var code int
                n, err := fmt.Sscanf(strings.ToLower(match), "&#x%x;", &code)
                if err == nil && n == 1 && code > 0 && code <= 0x10FFFF {
                        entitiesDecoded++
                        return string(rune(code))
                }
                return match
        })

        multipleSpacesPattern := regexp.MustCompile(`[ \t]+`)
        text = multipleSpacesPattern.ReplaceAllString(text, " ")

        multipleNewlinesPattern := regexp.MustCompile(`\n{3,}`)
        text = multipleNewlinesPattern.ReplaceAllString(text, "\n\n")

        return text, entitiesDecoded
}

func printResults(inputPath, outputPath string, stats *CleaningStats, showDetails bool, targetOS string) {
        fmt.Println("\n" + strings.Repeat("=", 70))
        fmt.Println("FILE CLEANING REPORT")
        fmt.Println(strings.Repeat("=", 70))

        fmt.Printf("\nFiles:\n")
        fmt.Printf("   Input:  %s\n", inputPath)
        fmt.Printf("   Output: %s\n", outputPath)

        fmt.Printf("\nConfiguration:\n")
        osName := targetOS
        lineEnding := "LF (\\n)"
        if targetOS == "windows" {
                osName = "Windows"
                lineEnding = "CRLF (\\r\\n)"
        } else if targetOS == "mac9" {
                osName = "Classic Mac OS"
                lineEnding = "CR (\\r)"
        } else {
                osName = "Unix/Linux/macOS"
                lineEnding = "LF (\\n)"
        }
        fmt.Printf("   Target OS:              %s\n", osName)
        fmt.Printf("   Line ending format:     %s\n", lineEnding)

        if stats.FormatDetected != "" {
                fmt.Printf("   Detected format:        %s\n", stats.FormatDetected)
        }
        if stats.MarkdownStripped {
                fmt.Printf("   Markdown stripped:      Yes\n")
        }
        if stats.HTMLStripped {
                fmt.Printf("   HTML stripped:          Yes\n")
                if stats.HTMLEntitiesDecoded > 0 {
                        fmt.Printf("   HTML entities decoded:  %d\n", stats.HTMLEntitiesDecoded)
                }
        }

        fmt.Printf("\nProcessing Statistics:\n")
        fmt.Printf("   Lines processed:        %d\n", stats.LinesProcessed)
        if stats.LinesWithIssues > 0 {
                fmt.Printf("   Lines with issues:      %d\n", stats.LinesWithIssues)
        }
        if stats.LineEndingsConverted > 0 {
                fmt.Printf("   Line endings converted: %d\n", stats.LineEndingsConverted)
        }
        fmt.Printf("   Total characters:       %d\n", stats.TotalChars)

        fmt.Printf("\nCharacter Removal Summary:\n")
        if stats.RemovedChars == 0 {
                fmt.Printf("   No invalid characters found - file is clean!\n")
        } else {
                fmt.Printf("   Total removed:        %d characters\n", stats.RemovedChars)
                if stats.ZeroWidthRemoved > 0 {
                        fmt.Printf("   Zero-width chars:     %d\n", stats.ZeroWidthRemoved)
                }
                if stats.ControlCharsRemoved > 0 {
                        fmt.Printf("   Control chars:        %d\n", stats.ControlCharsRemoved)
                }
                if stats.NonASCIIRemoved > 0 {
                        fmt.Printf("   Non-ASCII chars:      %d\n", stats.NonASCIIRemoved)
                }

                if stats.TotalChars > 0 {
                        percentage := float64(stats.RemovedChars) / float64(stats.TotalChars) * 100
                        fmt.Printf("\n   Removal rate: %.2f%% of total characters\n", percentage)
                }
        }

        if showDetails && len(stats.RemovedCharDetails) > 0 {
                fmt.Printf("\nDetailed Character Breakdown:\n")
                fmt.Println(strings.Repeat("-", 70))

                for char, count := range stats.RemovedCharDetails {
                        desc := charDescriptions[char]
                        if desc == "" {
                                if unicode.IsPrint(char) {
                                        desc = fmt.Sprintf("Character '%c'", char)
                                } else if unicode.IsControl(char) {
                                        desc = fmt.Sprintf("Control character (U+%04X)", char)
                                } else {
                                        desc = fmt.Sprintf("Non-printable (U+%04X)", char)
                                }
                        }

                        fmt.Printf("   U+%04X  %-40s  %d occurrence(s)\n", char, desc, count)
                }
                fmt.Println(strings.Repeat("-", 70))
        }

        fmt.Println("\n" + strings.Repeat("=", 70))
        if stats.RemovedChars > 0 || stats.LineEndingsConverted > 0 || stats.MarkdownStripped || stats.HTMLStripped {
                fmt.Println("File cleaned successfully!")
        } else {
                fmt.Println("File processed - no changes needed!")
        }
        fmt.Println(strings.Repeat("=", 70))
}

func cleanFile(inputPath, outputPath string, options CleaningOptions, verbose bool) (*CleaningStats, error) {
        contentBytes, err := os.ReadFile(inputPath)
        if err != nil {
                return nil, fmt.Errorf("could not read input file: %w", err)
        }

        content := string(contentBytes)

        stats := &CleaningStats{
                RemovedCharDetails: make(map[rune]int),
        }

        if options.StripFormat != "" {
                detectedFormat := detectFileFormat(content)
                stats.FormatDetected = detectedFormat

                if verbose {
                        fmt.Printf("Detected format: %s\n", detectedFormat)
                }

                if options.StripFormat == "markdown" {
                        if detectedFormat != "markdown" {
                                return nil, fmt.Errorf("file does not appear to be Markdown (detected: %s)", detectedFormat)
                        }
                        if verbose {
                                fmt.Println("Stripping Markdown formatting...")
                        }
                        content = stripMarkdown(content)
                        stats.MarkdownStripped = true
                } else if options.StripFormat == "html" {
                        if detectedFormat != "html" {
                                return nil, fmt.Errorf("file does not appear to be HTML (detected: %s)", detectedFormat)
                        }
                        if verbose {
                                fmt.Println("Stripping HTML tags and decoding entities...")
                        }
                        var entitiesDecoded int
                        content, entitiesDecoded = stripHTML(content)
                        stats.HTMLStripped = true
                        stats.HTMLEntitiesDecoded = entitiesDecoded
                }
        }

        outFile, err := os.Create(outputPath)
        if err != nil {
                return nil, fmt.Errorf("could not create output file: %w", err)
        }
        defer outFile.Close()

        writer := bufio.NewWriter(outFile)
        defer writer.Flush()

        lineNum := 0
        targetLineEnding := getLineEnding(options.TargetOS)
        lines := strings.Split(content, "\n")

        for i, line := range lines {
                lineNum++
                stats.LinesProcessed++

                if i < len(lines)-1 {
                        line += "\n"
                } else if len(contentBytes) > 0 && contentBytes[len(contentBytes)-1] == '\n' {
                        line += "\n"
                }

                cleanedLine, lineStats := cleanString(line, options)

                stats.TotalChars += lineStats.TotalChars
                stats.RemovedChars += lineStats.RemovedChars
                stats.NonASCIIRemoved += lineStats.NonASCIIRemoved
                stats.ControlCharsRemoved += lineStats.ControlCharsRemoved
                stats.ZeroWidthRemoved += lineStats.ZeroWidthRemoved

                for char, count := range lineStats.RemovedCharDetails {
                        stats.RemovedCharDetails[char] += count
                }

                if lineStats.RemovedChars > 0 {
                        stats.LinesWithIssues++
                }

                cleanedLine, converted := normalizeLineEndings(cleanedLine, targetLineEnding)
                if converted {
                        stats.LineEndingsConverted++
                }

                if verbose && lineStats.RemovedChars > 0 {
                        fmt.Printf("Line %d: Removed %d invalid character(s) ", lineNum, lineStats.RemovedChars)
                        fmt.Printf("[ZW:%d, Ctrl:%d, Non-ASCII:%d]\n",
                                lineStats.ZeroWidthRemoved,
                                lineStats.ControlCharsRemoved,
                                lineStats.NonASCIIRemoved)
                }

                if _, err := writer.WriteString(cleanedLine); err != nil {
                        return nil, fmt.Errorf("error writing to output: %w", err)
                }
        }

        if err := writer.Flush(); err != nil {
                return nil, fmt.Errorf("error flushing output: %w", err)
        }

        return stats, nil
}

func normalizeLineEndings(line, targetEnding string) (string, bool) {
        originalLine := line
        converted := false

        if strings.Contains(line, "\r\n") {
                if targetEnding != "\r\n" {
                        line = strings.ReplaceAll(line, "\r\n", targetEnding)
                        converted = true
                }
        } else if strings.Contains(line, "\n") {
                if targetEnding != "\n" {
                        line = strings.ReplaceAll(line, "\n", targetEnding)
                        converted = true
                }
        } else if strings.Contains(line, "\r") {
                if targetEnding != "\r" {
                        line = strings.ReplaceAll(line, "\r", targetEnding)
                        converted = true
                }
        }

        return line, converted && (line != originalLine)
}

func cleanString(s string, options CleaningOptions) (string, *CleaningStats) {
        stats := &CleaningStats{
                RemovedCharDetails: make(map[rune]int),
        }
        var result strings.Builder
        result.Grow(len(s))

        if len(s) == 0 {
                return s, stats
        }

        runes := []rune(s)

        startIdx := 0
        if options.RemoveBOM && len(runes) > 0 && runes[0] == '\uFEFF' {
                startIdx = 1
                stats.RemovedChars++
                stats.ZeroWidthRemoved++
                stats.TotalChars++
                stats.RemovedCharDetails['\uFEFF']++
        }

        for i := startIdx; i < len(runes); i++ {
                r := runes[i]
                stats.TotalChars++
                shouldKeep := true

                if options.RemoveZeroWidth && isZeroWidth(r) {
                        stats.RemovedChars++
                        stats.ZeroWidthRemoved++
                        stats.RemovedCharDetails[r]++
                        shouldKeep = false
                        continue
                }

                if options.RemoveNonASCII && r > 127 {
                        if r == '\n' || r == '\r' {
                                result.WriteRune(r)
                                continue
                        }
                        stats.RemovedChars++
                        stats.NonASCIIRemoved++
                        stats.RemovedCharDetails[r]++
                        shouldKeep = false
                        continue
                }

                if options.RemoveControlChars && unicode.IsControl(r) {
                        if r == '\n' || r == '\r' || r == '\t' {
                                result.WriteRune(r)
                                continue
                        }
                        stats.RemovedChars++
                        stats.ControlCharsRemoved++
                        stats.RemovedCharDetails[r]++
                        shouldKeep = false
                        continue
                }

                if options.NormalizeWhitespace && unicode.IsSpace(r) {
                        if (r == '\n' || r == '\r') && options.PreserveNewlines {
                                result.WriteRune(r)
                                continue
                        }
                        if (r == '\n' || r == '\r') && !options.PreserveNewlines {
                                continue
                        }
                        if r != ' ' {
                                result.WriteRune(' ')
                                continue
                        }
                }

                if shouldKeep {
                        result.WriteRune(r)
                }
        }

        return result.String(), stats
}

func isZeroWidth(r rune) bool {
        for _, zw := range zeroWidthChars {
                if r == zw {
                        return true
                }
        }
        return false
}

func copyFile(src, dst string) error {
        sourceInfo, err := os.Stat(src)
        if err != nil {
                if os.IsNotExist(err) {
                        return fmt.Errorf("source file does not exist: %s", src)
                }
                return fmt.Errorf("could not stat source file: %w", err)
        }

        if !sourceInfo.Mode().IsRegular() {
                return fmt.Errorf("source is not a regular file: %s", src)
        }

        sourceFile, err := os.Open(src)
        if err != nil {
                return fmt.Errorf("could not open source file: %w", err)
        }
        defer sourceFile.Close()

        destFile, err := os.Create(dst)
        if err != nil {
                return fmt.Errorf("could not create destination file: %w", err)
        }
        defer destFile.Close()

        bytesWritten, err := io.Copy(destFile, sourceFile)
        if err != nil {
                return fmt.Errorf("error copying file: %w", err)
        }

        if bytesWritten != sourceInfo.Size() {
                return fmt.Errorf("incomplete copy: wrote %d bytes, expected %d", bytesWritten, sourceInfo.Size())
        }

        if err := destFile.Sync(); err != nil {
                return fmt.Errorf("error syncing destination file: %w", err)
        }

        return nil
