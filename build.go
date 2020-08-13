package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	scheme "github.com/buildings-for-people/coding_scheme_object"
	scheme_pkg "github.com/buildings-for-people/coding_scheme_object"
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

func linksAreConsistent(ln string, scheme *scheme_pkg.Scheme) error {

	// See if we are expecting some sort of link.
	// if not, then it IS consistent.
	openSquare := strings.Index(ln, "[")
	if openSquare < 0 {
		return nil
	}

	// Otherwise, remove anything before the square bracket... the link
	// might be in parentheses
	ln = ln[openSquare+1:]
	// Then, we expect to find a closeSquare
	closeSquare := strings.Index(ln, "]")
	if closeSquare < 0 {
		return errors.New("expecting ']' in link")
	}
	// Then, an open round
	openRound := strings.Index(ln, "(")
	if openRound < 0 {
		return errors.New("expecting '(' in link")
	}

	// Then, an closed round
	closedRound := strings.Index(ln, ")")
	if closedRound < 0 {
		return errors.New("expecting ')' in link")
	}

	// If all that is there, check consistency of link.
	afterLink := ln[closedRound+1:]
	ln = ln[openRound+1 : closedRound]

	object := strings.Split(ln, "&")

	for _, component := range object {
		if strings.HasPrefix(component, "layer=") {
			layer := component[6:]
			layerName := scheme_pkg.IDToTxt(layer)

			if !scheme.HasLayer(layerName) {
				return fmt.Errorf("link leading to inexistent layer '%s'", layerName)
			}
		} else if strings.HasPrefix(component, "code=") {
			code := component[5:]
			codeName := scheme_pkg.IDToTxt(code)
			if !scheme.HasCode(codeName) {
				return fmt.Errorf("link leading to inexistent code '%s'", code)
			}

		} else if strings.HasPrefix(component, "http") {
			continue
		} else {
			return fmt.Errorf("incorrectly formatted link '%s'", ln)
		}

		/*
			// LINKS TO DOMAINS ARE NOT ALLOWED
			else if strings.HasPrefix(component, "domain=") {
				domain := component[7:]
				domainName := scheme_pkg.IDToTxt(domain)
				if !scheme.IsValidDomain(domainName) {
					return fmt.Errorf("link leading to inexistent domain '%s'", domain)
				}
			}
		*/
	}

	return linksAreConsistent(afterLink, scheme) //domains, layers, codes)

}

func checkDescription(filename string, scheme *scheme_pkg.Scheme) string { //domains *[]string, layers *[]string, codes *[]string) string {
	// Open file
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		abort("build.go", 87, err.Error())
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

			expectedFileName := scheme_pkg.TxtToFilename(ln)
			expectedFileName = expectedFileName[0 : len(expectedFileName)-3]

			// compare name of file with that line.
			aux := strings.Split(ln, "/")
			justName := aux[len(aux)-1]
			if expectedFileName != scheme_pkg.TxtToID(justName) {
				abort(filename, lineCount, fmt.Sprintf("Title ('%s') does not match filename ('%s')", expectedFileName, justName))
			}
			continue
		}

		d += (ln + "\n")

		// Check consistency of link.
		err := linksAreConsistent(ln, scheme)
		if err != nil {
			abort(filename, lineCount, err.Error())
		}

	}
	if lineCount < 1 {
		warn(fmt.Sprintf("File '%s' appears to be empty.", filename))
	}

	return formatDescription(d)

}

func main() {

	scheme := scheme.NewStandardScheme()

	// Go through each domain, checking that
	// files exist (warning if they do not)
	// and taking note of all the layers and codes found in such files
	domainsDir := "./domains"
	domainFiles, err := listMDFiles(domainsDir)
	if err != nil {
		abort("build.go", 289, err.Error())
	}

	for _, filename := range domainFiles {
		domainName := scheme_pkg.FilenameToTxt(filename)
		if !scheme.IsValidDomain(domainName) {
			abort("build.go", 193, fmt.Sprintf("File './%s' (domain '%s') was not expected", filename, domainName))
		}
		fullPath := fmt.Sprintf("%s/%s", domainsDir, filename)
		err := scheme.ReadDomainFile(fullPath, true)
		if err != nil {
			abort("build.go", 197, err.Error())
		}
	}

	// Create output directory if it does not exist.
	outdir := "./dist"
	if _, err := os.Stat(outdir); os.IsNotExist(err) {
		os.Mkdir(outdir, 0700)
	}

	// print the scheme as JSON
	j, e := json.Marshal(scheme)
	if e != nil {
		abort("build.go", 207, e.Error())
	}
	schemeFile := fmt.Sprintf("%s/scheme.json", outdir)
	e = ioutil.WriteFile(schemeFile, j, 0644)
	if err != nil {
		abort("build.go", 212, e.Error())
	}

	// Go through code files, checking that there are no files
	// that do not belong;
	codesDescriptions := make(map[string]string)
	codesDir := "./codes"

	codeFiles, err := listMDFiles(codesDir)
	if err != nil {
		abort("build.go", 163, err.Error())
	}

	for _, filename := range codeFiles {
		codeName := scheme_pkg.FilenameToTxt(filename)

		// Check if we were expecting this file
		if !scheme.HasCode(codeName) {
			warn(fmt.Sprintf("File '%s/%s' was not expected (codeName '%s')", codesDir, filename, codeName))
		}

		// check content of the file
		html := checkDescription(fmt.Sprintf("%s/%s", codesDir, filename), &scheme)
		codesDescriptions[scheme_pkg.TxtToID(codeName)] = html
	}

	// Then check that all the codes have files.

	contains := func(slice []string, s string) bool {
		for _, v := range slice {
			if v == s {
				return true
			}
		}
		return false
	}

	// Print the codes description as JSON
	j, e = json.Marshal(codesDescriptions)
	if e != nil {
		abort("build.go", 337, e.Error())
	}
	codesFile := fmt.Sprintf("%s/codes.json", outdir)
	e = ioutil.WriteFile(codesFile, j, 0644)
	if err != nil {
		abort("build.go", 342, e.Error())
	}

	////////////
	// Now, the same but with layers
	///////////

	layersDescriptions := make(map[string]string)
	layersDir := "./layers"
	layerFiles, err := listMDFiles(layersDir)
	if err != nil {
		abort("build.go", 350, err.Error())
	}

	for _, filename := range layerFiles {
		layerName := scheme_pkg.FilenameToTxt(filename)

		// Check if we were expecting this file
		if !scheme.HasLayer(layerName) {
			warn(fmt.Sprintf("File '%s/%s' was not expected (name '%s')", layersDir, filename, layerName))
		}

		html := checkDescription(fmt.Sprintf("%s/%s", layersDir, filename), &scheme) // &domains, &foundLayers, &foundCodes)
		layersDescriptions[scheme_pkg.TxtToID(layerName)] = html

	}

	// Print the layer description as JSON
	j, e = json.Marshal(layersDescriptions)
	if e != nil {
		abort("build.go", 369, e.Error())
	}
	layersFile := fmt.Sprintf("%s/layers.json", outdir)
	e = ioutil.WriteFile(layersFile, j, 0644)
	if err != nil {
		abort("build.go", 374, e.Error())
	}

	// Check which files are not there.

	warnedCodes := []string{}
	warnedLayers := []string{}
	for _, layer := range scheme.Layers {

		layerFileName := fmt.Sprintf("%s/%s", layersDir, scheme_pkg.TxtToFilename(layer.Name))

		if !fileExists(layerFileName) {
			warn(fmt.Sprintf("File '%s' is not there", layerFileName))
			if !contains(warnedLayers, layer.Name) {
				warnedCodes = append(warnedCodes, layer.Name)
			}
		}
		for _, code := range layer.Codes {
			filename := fmt.Sprintf("%s/%s", codesDir, scheme_pkg.TxtToFilename(code.Name))

			if !fileExists(filename) {
				warn(fmt.Sprintf("File '%s' is not there", filename))
				if !contains(warnedCodes, code.Name) {
					warnedCodes = append(warnedCodes, code.Name)
				}
			}
		}
	}

}
