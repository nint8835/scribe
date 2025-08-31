package embedding

import (
	"fmt"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/pipelines"
)

var Pipeline *pipelines.FeatureExtractionPipeline

func Initialize() error {
	session, err := hugot.NewGoSession()
	if err != nil {
		return fmt.Errorf("failed to create Hugot session: %w", err)
	}

	downloadOptions := hugot.NewDownloadOptions()
	downloadOptions.OnnxFilePath = "onnx/model.onnx"
	modelPath, err := hugot.DownloadModel(
		"sentence-transformers/all-MiniLM-L6-v2",
		"./models/",
		downloadOptions,
	)
	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}

	config := hugot.FeatureExtractionConfig{
		ModelPath: modelPath,
		Name:      "embeddingPipeline",
	}

	embeddingPipeline, err := hugot.NewPipeline(session, config)
	if err != nil {
		return fmt.Errorf("failed to create embedding pipeline: %w", err)
	}

	Pipeline = embeddingPipeline
	return nil
}

func EmbedQuote(text string) ([]byte, error) {
	//TODO: Preprocess text to remove mentions
	result, err := Pipeline.RunPipeline([]string{text})
	if err != nil {
		return []byte{}, fmt.Errorf("failed to run embedding pipeline: %w", err)
	}

	encodedEmbedding, err := sqlite_vec.SerializeFloat32(result.Embeddings[0])
	if err != nil {
		return []byte{}, fmt.Errorf("failed to serialize embedding: %w", err)
	}

	return encodedEmbedding, nil
}
