package embedding

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/knights-analytics/hugot"
	_ "github.com/mattn/go-sqlite3"
)

func TestPreprocessQuote(t *testing.T) {
	t.Parallel()

	text := "<@178958252820791296>: First line\nPlain line\n<@123>: Last line"
	want := "First line\nPlain line\nLast line"

	if got := preprocessQuote(text); got != want {
		t.Fatalf("preprocessQuote() = %q, want %q", got, want)
	}
}

func TestEmbeddingCacheReturnsCopies(t *testing.T) {
	t.Parallel()

	cache := newEmbeddingCache(2)
	cache.add("quote", []byte{1, 2, 3})

	first, ok := cache.get("quote")
	if !ok {
		t.Fatal("expected cached embedding")
	}
	first[0] = 9

	second, ok := cache.get("quote")
	if !ok {
		t.Fatal("expected cached embedding")
	}
	if !bytes.Equal(second, []byte{1, 2, 3}) {
		t.Fatalf("cached embedding was mutated: got %v", second)
	}
}

func TestEmbeddingCacheEvictsLeastRecentlyUsed(t *testing.T) {
	t.Parallel()

	cache := newEmbeddingCache(2)
	cache.add("first", []byte{1})
	cache.add("second", []byte{2})

	// Accessing first makes second the least recently used entry.
	if _, ok := cache.get("first"); !ok {
		t.Fatal("expected first embedding to be cached")
	}
	cache.add("third", []byte{3})

	if _, ok := cache.get("second"); ok {
		t.Fatal("expected least recently used embedding to be evicted")
	}
	if _, ok := cache.get("first"); !ok {
		t.Fatal("expected recently used embedding to remain cached")
	}
	if _, ok := cache.get("third"); !ok {
		t.Fatal("expected new embedding to be cached")
	}
}

func TestEmbedQuoteUsesPreprocessedTextAsCacheKey(t *testing.T) {
	const preprocessedText = "unique cached quote for EmbedQuote test"
	want := []byte{1, 2, 3}
	cache.add(preprocessedText, want)

	got, err := EmbedQuote(context.Background(), "<@123>: "+preprocessedText)
	if err != nil {
		t.Fatalf("EmbedQuote() returned an error: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("EmbedQuote() = %v, want %v", got, want)
	}
}

func BenchmarkEmbedQuoteRepeated(b *testing.B) {
	ctx := context.Background()
	modelPath, err := filepath.Abs("../../models/sentence-transformers_all-MiniLM-L6-v2")
	if err != nil {
		b.Fatalf("failed to resolve model path: %v", err)
	}
	if _, err := os.Stat(modelPath); err != nil {
		b.Skipf("local embedding model is unavailable: %v", err)
	}

	session, err := hugot.NewGoSession(ctx)
	if err != nil {
		b.Fatalf("NewGoSession() returned an error: %v", err)
	}
	b.Cleanup(func() {
		if err := session.Destroy(); err != nil {
			b.Errorf("Destroy() returned an error: %v", err)
		}
	})

	Pipeline, err = hugot.NewPipeline(session, hugot.FeatureExtractionConfig{
		ModelPath: modelPath,
		Name:      "embeddingBenchmarkPipeline",
	})
	if err != nil {
		b.Fatalf("NewPipeline() returned an error: %v", err)
	}

	const text = "The same quote is embedded repeatedly during semantic matching."
	preprocessedText := preprocessQuote(text)

	b.Run("BeforeWithoutCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := embedQuoteUncached(ctx, preprocessedText); err != nil {
				b.Fatalf("embedQuoteUncached() returned an error: %v", err)
			}
		}
	})

	b.Run("AfterWithCache", func(b *testing.B) {
		cache.clear()
		if _, err := EmbedQuote(ctx, text); err != nil {
			b.Fatalf("EmbedQuote() returned an error: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := EmbedQuote(ctx, text); err != nil {
				b.Fatalf("EmbedQuote() returned an error: %v", err)
			}
		}
	})
}
