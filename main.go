/*
 * Unlocks PDF files, tries to decrypt encrypted documents with the given password,
 * if that fails it tries an empty password as best effort.
 *
 * Run as: go run pdf_unlock.go input.pdf <password> output.pdf
 */

package main

import (
	"fmt"
	"os"

	pdf "github.com/unidoc/unidoc/pdf/model"
	"bufio"
    "github.com/slok/gospinner"
)

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func main() {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: go run pdf_unlock.go input.pdf <password_file> output.pdf\n")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	passwordFile := os.Args[2]
	outputPath := os.Args[3]

	s, err := gospinner.NewSpinner(gospinner.Dots2)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

    s.Start("Reading password file")
    //s.SetMessage(fmt.Sprintf("Reading password file"))
	lines, err := readLines(passwordFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
    s.Succeed()

    s.Start("Trying to crack PDF")
    total := len(lines)
	for i := 1; i < total; i++ {
        pos := float64(i)/float64(total)*100
		err := unlockPdf(inputPath, outputPath, lines[i])
		if err != nil {
            s.SetMessage(fmt.Sprintf("Working: %.3f %%", pos))
		} else {
			fmt.Printf("Complete. Password: %s see output file: %s\n", lines[i], outputPath)
            s.Succeed()
			os.Exit(0)
		}
	}
    s.Succeed()

}

func unlockPdf(inputPath string, outputPath string, password string) error {
	pdfWriter := pdf.NewPdfWriter()

	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	defer f.Close()

	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		return err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return err
	}

	// Try decrypting both with given password and an empty one if that fails.
	if isEncrypted {
		auth, err := pdfReader.Decrypt([]byte(password))
		if err != nil {
			return err
		}
		if !auth {
			return fmt.Errorf("Wrong password\n")
		}
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return err
		}

		err = pdfWriter.AddPage(page)
		if err != nil {
			return err
		}
	}

	fWrite, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer fWrite.Close()

	err = pdfWriter.Write(fWrite)
	if err != nil {
		return err
	}

	return nil
}

