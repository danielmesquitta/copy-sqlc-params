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

	// Prepare to write to the output file
	outFile, err := os.Create(outputFilePath)
	if err != nil {
		return "", fmt.Errorf("error creating output file: %w\n", err)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	// Write package declaration to the output file
	if _, err := writer.WriteString(fmt.Sprintf("package %s\n\n", packageName)); err != nil {
		return "", fmt.Errorf("error writing to file: %w\n", err)
	}

	// Check for models.go file and copy its imports
	modelsFile := filepath.Join(inputFolder, "models.go")
	if _, err := os.Stat(modelsFile); err == nil {
		if err := copyImports(modelsFile, writer); err != nil {
			return "", fmt.Errorf("error copying imports: %w\n", err)
		}
	}

	// Regular expressions to match struct definitions and names ending with Params
	structStartRegex := regexp.MustCompile(`^type (\w+Params) struct {`)
	structEndRegex := regexp.MustCompile(`^}`)

	// Walk through the files in the input folder
	err = filepath.Walk(
		inputFolder,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Process only .go files
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
				if err := processFile(path, structStartRegex, structEndRegex, writer); err != nil {
					return fmt.Errorf("error processing file: %w\n", err)
				}
			}
			return nil
		},
	)

	if err != nil {
		return "", fmt.Errorf("error reading folder: %w\n", err)
	}

	if err := formatGoFile(outputFilePath); err != nil {
		return "", fmt.Errorf("error formatting file: %w\n", err)
	}

	return outputFilePath, nil
}

func copyImports(filePath string, writer *bufio.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w\n", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	importsStarted := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "import (") {
			importsStarted = true
			if _, err := writer.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("error writing to file: %w\n", err)
			}
			continue
		}

		if importsStarted {
			if _, err := writer.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("error writing to file: %w\n", err)
			}
			if strings.HasPrefix(line, ")") {
				break
			}
		}
	}

	if importsStarted {
		if _, err := writer.WriteString("\n"); err != nil {
			return fmt.Errorf("error writing to file: %w\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %v\n", filePath, err)
	}

	return nil
}

func processFile(
	filePath string,
	structStartRegex, structEndRegex *regexp.Regexp,
	writer *bufio.Writer,
) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %v\n", filePath, err)
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
				// Write the captured struct to the output file
				if _, err := writer.WriteString(strings.Join(buffer, "\n") + "\n\n"); err != nil {
					return fmt.Errorf("error writing to file: %w\n", err)
				}
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
		return fmt.Errorf("error reading file %s: %v\n", filePath, err)
	}

	return nil
}

func formatGoFile(filePath string) error {
	command := exec.Command("gofmt", "-w", filePath)
	return command.Run()
}
