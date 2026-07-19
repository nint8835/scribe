package embedding

import (
	"container/list"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/pipelines"
)

const embeddingCacheCapacity = 1024

var Pipeline *pipelines.FeatureExtractionPipeline

type cacheEntry struct {
	text      string
	embedding []byte
}

type embeddingCache struct {
	mu       sync.Mutex
	capacity int
	entries  map[string]*list.Element
	order    *list.List
}

func newEmbeddingCache(capacity int) *embeddingCache {
	return &embeddingCache{
		capacity: capacity,
		entries:  make(map[string]*list.Element, capacity),
		order:    list.New(),
	}
}

func (c *embeddingCache) get(text string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, ok := c.entries[text]
	if !ok {
		return nil, false
	}

	c.order.MoveToFront(element)
	entry := element.Value.(*cacheEntry)
	return append([]byte(nil), entry.embedding...), true
}

func (c *embeddingCache) add(text string, encodedEmbedding []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.entries[text]; ok {
		entry := element.Value.(*cacheEntry)
		entry.embedding = append(entry.embedding[:0], encodedEmbedding...)
		c.order.MoveToFront(element)
		return
	}

	entry := &cacheEntry{
		text:      text,
		embedding: append([]byte(nil), encodedEmbedding...),
	}
	c.entries[text] = c.order.PushFront(entry)

	if c.order.Len() <= c.capacity {
		return
	}

	oldest := c.order.Back()
	delete(c.entries, oldest.Value.(*cacheEntry).text)
	c.order.Remove(oldest)
}

func (c *embeddingCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*list.Element, c.capacity)
	c.order.Init()
}

var cache = newEmbeddingCache(embeddingCacheCapacity)

func Initialize(ctx context.Context) error {
	slog.Debug("Initializing embedding support")

	session, err := hugot.NewGoSession(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Hugot session: %w", err)
	}

	modelsDir := "./models/"
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	downloadOptions := hugot.NewDownloadOptions()
	downloadOptions.OnnxFilePath = "onnx/model.onnx"

	slog.Debug("Downloading embedding model")
	modelPath, err := hugot.DownloadModel(
		ctx,
		"sentence-transformers/all-MiniLM-L6-v2",
		modelsDir,
		downloadOptions,
	)
	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}

	config := hugot.FeatureExtractionConfig{
		ModelPath: modelPath,
		Name:      "embeddingPipeline",
	}

	slog.Debug("Creating embedding pipeline", "modelPath", modelPath)
	embeddingPipeline, err := hugot.NewPipeline(session, config)
	if err != nil {
		return fmt.Errorf("failed to create embedding pipeline: %w", err)
	}

	slog.Debug("Pipeline created successfully")

	Pipeline = embeddingPipeline
	cache.clear()
	return nil
}

func preprocessQuote(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Remove Discord mention prefix pattern like "<@178958252820791296>: "
		if strings.HasPrefix(line, "<@") {
			_, after, ok := strings.Cut(line, ": ")
			if ok {
				lines[i] = after
			}
		}
	}
	return strings.Join(lines, "\n")
}

func EmbedQuote(ctx context.Context, text string) ([]byte, error) {
	preprocessedText := preprocessQuote(text)
	if text != preprocessedText {
		slog.Debug("Preprocessed quote text for embedding", "original", text, "preprocessed", preprocessedText)
	}

	if encodedEmbedding, ok := cache.get(preprocessedText); ok {
		return encodedEmbedding, nil
	}

	encodedEmbedding, err := embedQuoteUncached(ctx, preprocessedText)
	if err != nil {
		return []byte{}, err
	}

	cache.add(preprocessedText, encodedEmbedding)
	return encodedEmbedding, nil
}

func embedQuoteUncached(ctx context.Context, preprocessedText string) ([]byte, error) {
	result, err := Pipeline.RunPipeline(ctx, []string{preprocessedText})
	if err != nil {
		return []byte{}, fmt.Errorf("failed to run embedding pipeline: %w", err)
	}

	encodedEmbedding, err := sqlite_vec.SerializeFloat32(result.Embeddings[0])
	if err != nil {
		return []byte{}, fmt.Errorf("failed to serialize embedding: %w", err)
	}

	return encodedEmbedding, nil
}
