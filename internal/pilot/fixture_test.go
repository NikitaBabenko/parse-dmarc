package pilot

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"path/filepath"
	"testing"

	"github.com/dmarcguardhq/parse-dmarc/internal/parser"
)

func TestSeedOptInAndDoesNotOverwrite(t *testing.T) {
	if Seed("") == nil {
		t.Fatal("empty path accepted")
	}
	dir := t.TempDir()
	if err := Seed(dir); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "new", filename)
	first, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(first, []byte("example.test")) {
		t.Fatal("synthetic message missing")
	}
	if err := os.WriteFile(target, []byte("existing"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Seed(dir); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(target)
	if string(after) != "existing" {
		t.Fatal("overwrote existing mail")
	}
}

func TestSyntheticReportContent(t *testing.T) {
	email, err := mail.ReadMessage(bytes.NewReader(message))
	if err != nil {
		t.Fatal(err)
	}
	_, params, err := mime.ParseMediaType(email.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	parts := multipart.NewReader(email.Body, params["boundary"])
	for {
		part, err := parts.NextPart()
		if err == io.EOF {
			t.Fatal("missing XML attachment")
		}
		if err != nil {
			t.Fatal(err)
		}
		if part.FileName() != "example.test.xml" {
			continue
		}
		xml, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		report, err := parser.ParseReport(xml)
		if err != nil {
			t.Fatal(err)
		}
		if report.PolicyPublished.Domain != "example.test" || report.GetTotalMessages() != 10 || report.GetDMARCCompliantCount() != 8 || len(report.Records) != 2 {
			t.Fatal("unexpected synthetic report")
		}
		if report.Records[0].Row.SourceIP != "192.0.2.10" || report.Records[1].Row.SourceIP != "198.51.100.20" {
			t.Fatal("unexpected IP addresses")
		}
		return
	}
}
