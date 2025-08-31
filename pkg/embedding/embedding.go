package embedding

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/pipelines"
)

var Pipeline *pipelines.FeatureExtractionPipeline

func Initialize() error {
	slog.Debug("Initializing embedding support")

	session, err := hugot.NewGoSession()
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
	return nil
}

func EmbedQuote(text string) ([]byte, error) {
	// Preprocess text to remove mentions
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Remove Discord mention prefix pattern like "<@178958252820791296>: "
		if strings.HasPrefix(line, "<@") {
			colonIndex := strings.Index(line, ": ")
			if colonIndex != -1 {
				lines[i] = line[colonIndex+2:]
			}
		}
	}
	preprocessedText := strings.Join(lines, "\n")

	if text != preprocessedText {
		slog.Debug("Preprocessed quote text for embedding", "original", text, "preprocessed", preprocessedText)
	}

	result, err := Pipeline.RunPipeline([]string{preprocessedText})
	if err != nil {
		return []byte{}, fmt.Errorf("failed to run embedding pipeline: %w", err)
	}

	encodedEmbedding, err := sqlite_vec.SerializeFloat32(result.Embeddings[0])
	if err != nil {
		return []byte{}, fmt.Errorf("failed to serialize embedding: %w", err)
	}

	return encodedEmbedding, nil
}
