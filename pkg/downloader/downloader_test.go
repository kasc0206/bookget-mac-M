package downloader

import (
	"context"
	"net/http/cookiejar"
	"testing"
)

func TestVec2d(t *testing.T) {
	v := NewVec2d(100, 200)
	if v.Width() != 100 {
		t.Errorf("Width() = %d, want 100", v.Width())
	}
	if v.Height() != 200 {
		t.Errorf("Height() = %d, want 200", v.Height())
	}
}

func TestTileSizeFormat(t *testing.T) {
	if WidthHeight != 0 {
		t.Errorf("WidthHeight = %d, want 0", WidthHeight)
	}
	if Width != 1 {
		t.Errorf("Width = %d, want 1", Width)
	}
}

func TestQualityOrder(t *testing.T) {
	if len(qualityOrder) != 2 {
		t.Errorf("qualityOrder len = %d, want 2", len(qualityOrder))
	}
	if qualityOrder[0] != "default" {
		t.Errorf("qualityOrder[0] = %s, want default", qualityOrder[0])
	}
	if qualityOrder[1] != "native" {
		t.Errorf("qualityOrder[1] = %s, want native", qualityOrder[1])
	}
}

func TestFormatOrder(t *testing.T) {
	if len(formatOrder) != 2 {
		t.Errorf("formatOrder len = %d, want 2", len(formatOrder))
	}
	if formatOrder[0] != "jpg" {
		t.Errorf("formatOrder[0] = %s, want jpg", formatOrder[0])
	}
	if formatOrder[1] != "png" {
		t.Errorf("formatOrder[1] = %s, want png", formatOrder[1])
	}
}

func TestNewDownloadManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dm := NewDownloadManager(ctx, cancel, 8)
	if dm == nil {
		t.Fatal("NewDownloadManager returned nil")
	}
	if dm.maxConcurrent != 8 {
		t.Errorf("maxConcurrent = %d, want 8", dm.maxConcurrent)
	}
}

func TestNewDownloadManager_ClampMin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dm := NewDownloadManager(ctx, cancel, 0)
	if dm.maxConcurrent != 16 {
		t.Errorf("maxConcurrent should default to 16, got %d", dm.maxConcurrent)
	}
}

func TestNewDownloadManager_ClampMax(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dm := NewDownloadManager(ctx, cancel, 100)
	if dm.maxConcurrent != 100 {
		t.Errorf("maxConcurrent = %d, want 100 (unclamped)", dm.maxConcurrent)
	}
}

func TestAddTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dm := NewDownloadManager(ctx, cancel, 4)
	if len(dm.tasks) != 0 {
		t.Errorf("initial tasks = %d, want 0", len(dm.tasks))
	}

	dm.AddTask("http://example.com/file.jpg", "GET", nil, nil, "/tmp", "file.jpg", 2)
	if len(dm.tasks) != 1 {
		t.Errorf("tasks after add = %d, want 1", len(dm.tasks))
	}

	task := dm.tasks[0]
	if task.URL != "http://example.com/file.jpg" {
		t.Errorf("task.URL = %s, want http://example.com/file.jpg", task.URL)
	}
	if task.FileName != "file.jpg" {
		t.Errorf("task.FileName = %s, want file.jpg", task.FileName)
	}
}

func TestAddTask_ClampThreads(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dm := NewDownloadManager(ctx, cancel, 4)
	dm.AddTask("http://example.com/file.jpg", "GET", nil, nil, "/tmp", "file.jpg", 0)
	if dm.tasks[0].Threads != 1 {
		t.Errorf("Threads should default to 1, got %d", dm.tasks[0].Threads)
	}
}

func TestDownloadTask_HttpClient_Default(t *testing.T) {
	task := &DownloadTask{URL: "http://example.com/file.jpg"}
	client := task.httpClient()
	if client == nil {
		t.Fatal("httpClient() returned nil")
	}
	if client.Jar != nil {
		t.Error("expected nil Jar by default")
	}
}

func TestDownloadTask_HttpClient_WithJar(t *testing.T) {
	jar, _ := cookiejar.New(nil)
	task := &DownloadTask{
		URL: "http://example.com/file.jpg",
		Jar: jar,
	}
	client := task.httpClient()
	if client == nil {
		t.Fatal("httpClient() returned nil")
	}
	if client.Jar == nil {
		t.Error("expected non-nil Jar")
	}
	if client.Jar != jar {
		t.Error("expected same Jar instance")
	}
}

func TestConstants(t *testing.T) {
	if minFileSize != 1024 {
		t.Errorf("minFileSize = %d, want 1024", minFileSize)
	}
	if maxRetries != 3 {
		t.Errorf("maxRetries = %d, want 3", maxRetries)
	}
	if JPGQuality != 90 {
		t.Errorf("JPGQuality = %d, want 90", JPGQuality)
	}
}
