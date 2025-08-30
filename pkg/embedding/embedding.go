package embedding

import (
	"fmt"

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
