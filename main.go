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
	"log"
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

	s.Start("Opening password file")
	file, err := os.Open(passwordFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	s.Succeed()

	s.Start("Trying to crack PDF")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		s.SetMessage(fmt.Sprintf("Working: %s", line))
		err := unlockPdf(inputPath, outputPath, line)
		if err != nil {
		} else {
			fmt.Printf("Complete. Password: %s see output file: %s\n", line, outputPath)
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

