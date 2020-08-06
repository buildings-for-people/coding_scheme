package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gitlab.com/golang-commonmark/markdown"
)

func formatDescription(d string) string {
	md := markdown.New(markdown.XHTMLOutput(true))
	return md.RenderToString([]byte(d))
}

func abort(filename string, line int, msg string) {
	fmt.Println(fmt.Sprintf("FATAL ERROR [%s : ln %d]: %s", filename, line, msg))
	os.Exit(1)
}

func warn(msg string) {
	fmt.Println(fmt.Sprintf("WARNING: %s", msg))
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func listMDFiles(dirname string) ([]string, error) {

	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return []string{}, err
	}

	ret := []string{}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".md") {
			ret = append(ret, f.Name())
		}
	}
	return ret, nil
}

func toID(ln string) string {
	ln = strings.TrimSpace(ln)
	ln = strings.ToLower(ln)
	ln = strings.Join(strings.Split(ln, " "), "_")
	return ln
}

func toFilename(ln string) string {
	return fmt.Sprintf("%s.md", toID(ln))
}

func filenameToTxt(ln string) string {
	ln = idToTxt(ln)
	ln = ln[0 : len(ln)-3]
	return ln
}

func idToTxt(ln string) string {
	return strings.Join(strings.Split(ln, "_"), " ")
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func checkDescription(filename string, domains *[]string, layers *[]string, codes *[]string) {
	// Open file
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		abort("check.go", 75, err.Error())
	}

	d := ""
	// scan all lines
	lineCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCount++
		ln := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(ln, "#") {
			// A title... check it matches the name of the file
			ln = ln[2:]
			ln = strings.TrimSpace(ln)

			expectedFileName := toFilename(ln)
			expectedFileName = expectedFileName[0 : len(expectedFileName)-3]

			// compare name of file with that line.
			aux := strings.Split(ln, "/")
			justName := aux[len(aux)-1]
			if expectedFileName != toID(justName) {
				abort(filename, lineCount, fmt.Sprintf("Title ('%s') does not match filename ('%s')", expectedFileName, justName))
			}
			continue
		}

		d += ln

		// if this is true, we are expecting some sort of link.
		openSquare := strings.Index(ln, "[")
		if openSquare >= 0 {
			// remove anything before the square bracket... the link
			// might be in parentheses
			ln = ln[openSquare+1:]
			// Then, we expect to find a closeSquare
			closeSquare := strings.Index(ln, "]")
			if closeSquare < 0 {
				abort(filename, lineCount, "Expecting ']' in link.")
			}
			// Then, an open round
			openRound := strings.Index(ln, "(")
			if openRound < 0 {
				abort(filename, lineCount, "Expecting '(' in link.")
			}

			// Then, an closed round
			closedRound := strings.Index(ln, ")")
			if closedRound < 0 {
				abort(filename, lineCount, "Expecting ')' in link.")
			}

			// If all that is there, check consistency of link.
			ln = ln[openRound+1 : closedRound]

			object := strings.Split(ln, "&")

			for _, component := range object {
				if strings.HasPrefix(component, "layer=") {
					layer := component[6:]
					if !contains(*layers, idToTxt(layer)) {
						warn(fmt.Sprintf("Link leading to inexistent layer... File '%s' line %d, leading to '%s'", filename, lineCount, layer))
					}
				} else if strings.HasPrefix(component, "code=") {
					code := component[5:]
					if !contains(*codes, idToTxt(code)) {
						warn(fmt.Sprintf("Link leading to inexistent code... File '%s' line %d, leading to '%s'", filename, lineCount, code))
					}

				} else if strings.HasPrefix(component, "domain=") {
					domain := component[7:]
					if !contains(*domains, idToTxt(domain)) {
						warn(fmt.Sprintf("Link leading to inexistent domain... File '%s' line %d, leading to '%s'", filename, lineCount, domain))
					}
				} else if strings.HasPrefix(component, "http") {
					continue
				} else {
					abort(filename, lineCount, fmt.Sprintf("Incorrectly formatted link '%s'", ln))
				}
			}
		}
	}
	if lineCount < 1 {
		warn(fmt.Sprintf("File '%s' appears to be empty.", filename))
	}
	// IF WE WANT TO CREATE NEW FILES

	// Now, process...
	filename = "./html/" + filename
	dirs := strings.Split(filename, "/")

	baseDir := "."
	for _, dir := range dirs {
		baseDir = fmt.Sprintf("%s/%s", baseDir, dir)
		if strings.HasSuffix(dir, ".md") {
			baseDir = baseDir[0 : len(baseDir)-3]

			// Write the file.
			html := formatDescription(d)

			err = ioutil.WriteFile(fmt.Sprintf("%s.html", baseDir), []byte(html), 0644)
			if err != nil {
				abort("check.go", 188, err.Error())
			}

		} else {
			if _, err := os.Stat(baseDir); os.IsNotExist(err) {
				os.Mkdir(baseDir, 0777)
			}
		}
	}

}

func checkDomainFile(filename string, foundLayers *[]string, foundCodes *[]string) {

	warnedCodes := make([]string, 0)
	warnedLayers := make([]string, 0)

	// Open file
	domainFile, err := os.Open(fmt.Sprintf("./domains/%s", filename))
	defer domainFile.Close()
	if err != nil {
		abort("check.go", 61, err.Error())
	}

	// scan all lines
	lineCount := 0
	scanner := bufio.NewScanner(domainFile)
	for scanner.Scan() {
		lineCount++
		// clean the line
		ln := strings.TrimSpace(scanner.Text())
		ln = strings.ToLower(ln)

		if ln == "" {
			continue
		} else if strings.HasPrefix(ln, "# ") {
			// Domain name...
			ln = ln[2:]
			ln = strings.TrimSpace(ln)

			expectedFileName := toFilename(ln)

			// compare name of file with that line.
			if expectedFileName != filename {
				abort(filename, lineCount, fmt.Sprintf("Domain name ('%s') does not match filename ('%s')", ln, filename))
			}

		} else if strings.HasPrefix(ln, "## ") {
			// Layer name...
			ln = ln[3:]
			ln = strings.TrimSpace(ln)

			if !contains(*foundLayers, ln) {
				*foundLayers = append(*foundLayers, ln)
			}

			// check that file is there.
			layerFileName := fmt.Sprintf("./layers/%s", toFilename(ln))

			if !fileExists(layerFileName) && !contains(warnedLayers, ln) {
				warnedLayers = append(warnedLayers, ln)
				warn(fmt.Sprintf("File '%s' is not there", layerFileName))
			}

		} else if strings.HasPrefix(ln, "* ") {
			// Code...
			ln = ln[2:]
			ln = strings.TrimSpace(ln)

			if !contains(*foundCodes, ln) {
				*foundCodes = append(*foundCodes, ln)
			}

			// check that there is a file about that.
			codeFileName := fmt.Sprintf("./codes/%s", toFilename(ln))

			if !fileExists(codeFileName) && !contains(warnedCodes, ln) {
				warnedCodes = append(warnedCodes, ln)
				warn(fmt.Sprintf("File '%s' is not there", codeFileName))
			}

		} else {
			abort(filename, lineCount, fmt.Sprintf("Unexpected content '%s'", ln))
		}

	} // end iterating file lines
}

func main() {

	var domains = []string{"acoustic", "air quality", "coolness", "daylight", "warmness"}

	foundLayers := make([]string, 0)
	foundCodes := make([]string, 0)

	// Go through each domain, checking that
	// files exist (warning if they do not)
	// and taking note of all the layers and codes
	// found in such files
	domainFiles, err := listMDFiles("./domains")
	if err != nil {
		abort("check.go", 142, err.Error())
	}

	for _, filename := range domainFiles {
		if !contains(domains, filenameToTxt(filename)) {
			warn(fmt.Sprintf("File './%s' was not expected", filename))
		}
		checkDomainFile(filename, &foundLayers, &foundCodes)
	}

	// Go through codes, checking that there are no files
	// that do not belong.
	codeFiles, err := listMDFiles("./codes")
	if err != nil {
		abort("check.go", 163, err.Error())
	}

	for _, filename := range codeFiles {
		codeName := filenameToTxt(filename)

		// Check if we were expecting this file
		if !contains(foundCodes, codeName) {
			warn(fmt.Sprintf("File './codes/%s' was not expected (code '%s')", filename, codeName))
		}

		// check content of the file
		checkDescription(fmt.Sprintf("./codes/%s", filename), &domains, &foundLayers, &foundCodes)
	}

	// Now, the same but with layers

	layerFiles, err := listMDFiles("./layers")
	if err != nil {
		abort("check.go", 181, err.Error())
	}

	for _, filename := range layerFiles {
		layerName := filenameToTxt(filename)

		// Check if we were expecting this file
		if !contains(foundLayers, layerName) {
			warn(fmt.Sprintf("File './layers/%s' was not expected (name '%s')", filename, layerName))
		}

		checkDescription(fmt.Sprintf("./layers/%s", filename), &domains, &foundLayers, &foundCodes)
	}

}
