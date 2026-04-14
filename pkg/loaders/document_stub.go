package loaders

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

func extractPDFText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("pdf open: %w", err)
	}
	defer f.Close()
	reader, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("pdf text: %w", err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("pdf read: %w", err)
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", fmt.Errorf("pdf text: no extractable text")
	}
	return text, nil
}

func extractDOCXText(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("docx open: %w", err)
	}
	defer r.Close()
	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}
		text, err := readDOCXDocumentXML(f)
		if err != nil {
			return "", err
		}
		if text == "" {
			return "", fmt.Errorf("docx text: no extractable text")
		}
		return text, nil
	}
	return "", fmt.Errorf("docx text: word/document.xml not found")
}

func readDOCXDocumentXML(file *zip.File) (string, error) {
	rc, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("docx xml open: %w", err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("docx xml read: %w", err)
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	var out strings.Builder
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("docx xml decode: %w", err)
		}
		switch se := tok.(type) {
		case xml.StartElement:
			if se.Name.Local == "p" && out.Len() > 0 {
				out.WriteString("\n")
			}
		case xml.CharData:
			text := strings.TrimSpace(string(se))
			if text == "" {
				continue
			}
			if out.Len() > 0 && !strings.HasSuffix(out.String(), "\n") {
				out.WriteString(" ")
			}
			out.WriteString(text)
		}
	}
	return strings.TrimSpace(out.String()), nil
}
