package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestServer tests the behavior of the static file serving server.
func TestServer(t *testing.T) {
	// Create a temporary directory to store test files
	tempDir := "./temp_test_dir"
	err := os.Mkdir(tempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after tests

	// Create some test files in the temporary directory
	testFiles := map[string]string{
		"index.html": "<html><body><h1>Test Page</h1></body></html>",
		"styles.css": "body { background-color: #fff; }",
		"script.js":  "console.log('Test JS');",
		"image.png":  "fakeimagecontent", // Pretend it's a PNG file
	}
	for fileName, content := range testFiles {
		filePath := tempDir + "/" + fileName
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", fileName, err)
		}
	}

	// Set up the file server to serve from the tempDir
	fs := http.FileServer(http.Dir(tempDir))
	http.Handle("/", fs)

	// Create an in-memory test server
	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	// Test Serving an Existing File (index.html)
	t.Run("Serving Existing File", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/index.html")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
		}

		// Check the response body content
		expectedBody := testFiles["index.html"]
		body := make([]byte, len(expectedBody))
		_, err = resp.Body.Read(body)
		if err != nil && err.Error() != "EOF" {
			t.Fatalf("Failed to read response body: %v", err)
		}
		if string(body) != expectedBody {
			t.Errorf("Expected body %q, got %q", expectedBody, string(body))
		}
	})

	// Test Handling Non-Existent File (404 Error)
	t.Run("Handling Non-Existent File", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/nonexistent.html")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 Not Found, got %d", resp.StatusCode)
		}
	})

	// Test Handling Different MIME Types
	t.Run("Managing Different MIME Types", func(t *testing.T) {
		// Test HTML file MIME type
		resp, err := http.Get(ts.URL + "/index.html")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
			t.Errorf("Expected Content-Type to be text/html, got %s", resp.Header.Get("Content-Type"))
		}

		// Test CSS file MIME type
		resp, err = http.Get(ts.URL + "/styles.css")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		if !strings.Contains(resp.Header.Get("Content-Type"), "text/css") {
			t.Errorf("Expected Content-Type to be text/css, got %s", resp.Header.Get("Content-Type"))
		}

		// Test JS file MIME type
		resp, err = http.Get(ts.URL + "/script.js")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		// Check that Content-Type contains 'text/javascript' (with or without charset)
		if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/javascript") {
			t.Errorf("Expected Content-Type to be text/javascript, got %s", resp.Header.Get("Content-Type"))
		}

		// Test PNG file MIME type
		resp, err = http.Get(ts.URL + "/image.png")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		defer resp.Body.Close()

		if !strings.Contains(resp.Header.Get("Content-Type"), "image/png") {
			t.Errorf("Expected Content-Type to be image/png, got %s", resp.Header.Get("Content-Type"))
		}
	})
}
