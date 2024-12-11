package usecase

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func CopySQLCParams(
	inputFolder, outputFolder, packageName string,
) (string, error) {
	if inputFolder == "" {
		return "", fmt.Errorf("input folder is required")
	}
	if outputFolder == "" {
		return "", fmt.Errorf("output folder is required")
	}

	if packageName == "" {
		packageName = filepath.Base(outputFolder)
	}

	outputFile := "params_gen.go"
	outputFilePath := filepath.Join(outputFolder, outputFile)

	var result strings.Builder

	// Write package declaration
	result.WriteString(
		fmt.Sprintf(
			"//nolint\n//go:build !codeanalysis\n// +build !codeanalysis\n\npackage %s\n\n",
			packageName,
		),
	)

	// Check for models.go file and copy its imports
	modelsFile := filepath.Join(inputFolder, "models.go")
	if _, err := os.Stat(modelsFile); err == nil {
		if err := copyImports(modelsFile, &result); err != nil {
			return "", fmt.Errorf("error copying imports: %w", err)
		}
	}

	// Regular expressions to match struct definitions and names ending with Params
	structStartRegex := regexp.MustCompile(`^type (\w+Params) struct {`)
	structEndRegex := regexp.MustCompile(`^}`)

	// Walk through the files in the input folder
	err := filepath.Walk(
		inputFolder,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Process only .go files
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
				if err := processFile(path, structStartRegex, structEndRegex, &result); err != nil {
					return fmt.Errorf("error processing file: %w", err)
				}
			}
			return nil
		},
	)

	if err != nil {
		return "", fmt.Errorf("error reading folder: %w", err)
	}

	// Write the collected content to the output file
	if err := os.MkdirAll(outputFolder, os.ModePerm); err != nil {
		return "", fmt.Errorf("error creating output directory: %w", err)
	}

	if err := os.WriteFile(outputFilePath, []byte(result.String()), 0644); err != nil {
		return "", fmt.Errorf("error writing to file: %w", err)
	}

	if err := formatGoFile(outputFilePath); err != nil {
		return "", fmt.Errorf("error formatting file: %w", err)
	}

	return outputFilePath, nil
}

func copyImports(filePath string, result *strings.Builder) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	importsStarted := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "import (") {
			importsStarted = true
			result.WriteString(line + "\n")
			continue
		}

		if importsStarted {
			result.WriteString(line + "\n")
			if strings.HasPrefix(line, ")") {
				break
			}
		}
	}

	if importsStarted {
		result.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	return nil
}

func processFile(
	filePath string,
	structStartRegex, structEndRegex *regexp.Regexp,
	result *strings.Builder,
) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %v", filePath, err)
	}
	defer file.Close()

	var capturing bool
	var buffer []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if capturing {
			buffer = append(buffer, line)
			if structEndRegex.MatchString(line) {
				// Write the captured struct to the result
				result.WriteString(strings.Join(buffer, "\n") + "\n\n")
				buffer = nil
				capturing = false
			}
			continue
		}

		if matches := structStartRegex.FindStringSubmatch(line); matches != nil {
			capturing = true
			buffer = append(buffer, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	return nil
}

func formatGoFile(filePath string) error {
	command := exec.Command("gofmt", "-w", filePath)
	return command.Run()
}
