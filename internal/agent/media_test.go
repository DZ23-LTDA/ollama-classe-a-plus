package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMediaManagerGeneratesArtifactsThroughHTTPSProvider(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/images/generations":
			png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 'p', 'n', 'g'}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{map[string]any{"b64_json": base64.StdEncoding.EncodeToString(png)}}})
		case "/audio/transcriptions":
			_ = json.NewEncoder(w).Encode(map[string]any{"text": "transcrição aprovada"})
		case "/chat/completions":
			_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": "OCR: texto visível"}}}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	manager, err := NewMediaManager(MediaProvider{Name: "test", BaseURL: server.URL, APIKey: "test-key", ImageModel: "image-test", TranscriptionModel: "stt-test"})
	if err != nil {
		t.Fatal(err)
	}
	manager.Client = server.Client()
	workspace := t.TempDir()
	image, err := manager.GenerateImage(context.Background(), workspace, "uma imagem de teste", "")
	if err != nil {
		t.Fatal(err)
	}
	if image.Artifact.SHA256 == "" || image.Artifact.Path == "" {
		t.Fatalf("invalid image artifact: %+v", image.Artifact)
	}
	input := filepath.Join(workspace, "input.wav")
	if err := os.WriteFile(input, []byte("wav"), 0o600); err != nil {
		t.Fatal(err)
	}
	transcript, err := manager.Transcribe(context.Background(), workspace, input, "")
	if err != nil || transcript.Text != "transcrição aprovada" {
		t.Fatalf("transcript=%+v err=%v", transcript, err)
	}
	inputImage := filepath.Join(workspace, "input.png")
	if err := os.WriteFile(inputImage, []byte("fake-png"), 0o600); err != nil {
		t.Fatal(err)
	}
	vision, err := manager.AnalyzeImage(context.Background(), workspace, inputImage, "Extraia o texto da imagem", "vision-test")
	if err != nil || vision.Text != "OCR: texto visível" {
		t.Fatalf("vision=%+v err=%v", vision, err)
	}
}

func TestGenerateToneProducesWAVArtifact(t *testing.T) {
	result, err := GenerateTone(t.TempDir(), 440, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 44 || string(data[:4]) != "RIFF" || result.MediaType != "audio/wav" {
		t.Fatalf("invalid wav: len=%d type=%s", len(data), result.MediaType)
	}
}

func TestMediaMaterializeRejectsRedirectAndInvalidMagic(t *testing.T) {
	if err := validateMediaContentType(".png", "text/html"); err == nil {
		t.Fatal("expected MIME mismatch rejection")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("not-a-png"))
	}))
	defer server.Close()
	manager, err := NewMediaManager(MediaProvider{Name: "local", BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.materializeEntry(context.Background(), t.TempDir(), "image", ".png", map[string]any{"url": server.URL + "/redirect"}); err == nil || !strings.Contains(err.Error(), "redirect") {
		t.Fatalf("expected redirect rejection, got %v", err)
	}
	if _, _, err := manager.materializeEntry(context.Background(), t.TempDir(), "image", ".png", map[string]any{"b64_json": base64.StdEncoding.EncodeToString([]byte("not-a-png"))}); err == nil || !strings.Contains(err.Error(), "PNG") {
		t.Fatalf("expected magic rejection, got %v", err)
	}
}

func TestMediaDownloadLimitsAndRejectsPrivateActualAddress(t *testing.T) {
	if _, err := readLimitedMediaBody(strings.NewReader("1234"), 3); err == nil {
		t.Fatal("expected media body limit rejection")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan struct{})
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			_ = conn.Close()
		}
		close(accepted)
	}()
	_, err = mediaDialContext(context.Background(), "tcp", listener.Addr().String())
	<-accepted
	if err == nil || !strings.Contains(err.Error(), "private") {
		t.Fatalf("expected private media dial rejection, got %v", err)
	}
}
