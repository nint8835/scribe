package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/goccy/go-graphviz"

	"github.com/nint8835/scribe/pkg/config"
	"github.com/nint8835/scribe/pkg/database"
	"github.com/nint8835/scribe/pkg/embedding"
)

func main() {
	err := config.Load()
	if err != nil {
		slog.Error("Error loading config", "error", err)
		os.Exit(1)
	}

	err = embedding.Initialize()
	if err != nil {
		slog.Error("Error initializing embedding", "error", err)
		os.Exit(1)
	}

	database.Initialize(config.Instance.DBPath)

	var maxQuoteId int64
	err = database.Instance.Model(&database.Quote{}).Select("MAX(id)").Scan(&maxQuoteId).Error
	if err != nil {
		slog.Error("Error counting quotes", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	g, err := graphviz.New(ctx)
	if err != nil {
		slog.Error("Error creating graphviz instance", "error", err)
		os.Exit(1)
	}

	graph, err := g.Graph()
	if err != nil {
		slog.Error("Error creating graph", "error", err)
		os.Exit(1)
	}

	// maxQuoteId = 200

	for i := int64(0); i < maxQuoteId; i++ {
		nodeName := fmt.Sprintf("%d", i)

		n1, err := graph.CreateNodeByName(nodeName)
		if err != nil {
			slog.Error("Error creating node", "error", err)
			os.Exit(1)
		}

		var quote database.Quote
		database.Instance.Model(database.Quote{}).First(&quote, i)

		encodedEmbedding, err := embedding.EmbedQuote(quote.Text)
		if err != nil {
			slog.Error("Error embedding quote", "error", err)
			os.Exit(1)
		}

		var neighbourWeights []struct {
			ID       int
			Distance float64
		}

		err = database.Instance.Raw(
			`SELECT
				rowid AS id,
				distance
			FROM quote_embeddings
			WHERE
				embedding MATCH ?
				AND rowid != ?
				AND k = 2
				AND distance >= 0.4
			ORDER BY distance`,
			encodedEmbedding,
			quote.Meta.ID,
			maxQuoteId,
		).Scan(&neighbourWeights).Error
		if err != nil {
			slog.Error("Error querying neighbours", "error", err)
			os.Exit(1)
		}

		for _, neighbour := range neighbourWeights {
			// slog.Debug("Distance", "distance", neighbour.Distance)

			neighbourNodeName := fmt.Sprintf("%d", neighbour.ID)
			n2, err := graph.CreateNodeByName(neighbourNodeName)
			if err != nil {
				slog.Error("Error creating neighbour node", "error", err)
				os.Exit(1)
			}

			edgeName := fmt.Sprintf("%d-%d", quote.Meta.ID, neighbour.ID)
			edge, err := graph.CreateEdgeByName(edgeName, n1, n2)
			if err != nil {
				slog.Error("Error creating edge", "error", err)
				os.Exit(1)
			}

			edge.SetWeight(neighbour.Distance)
			edge.SetLen((1 - neighbour.Distance) * 3)
		}
	}

	graph.SetLayout("neato")
	graph.SetOverlap(false)

	f, err := os.Create("graph.dot")
	if err != nil {
		slog.Error("Error creating dot file", "error", err)
		os.Exit(1)
	}
	defer f.Close()

	err = g.Render(ctx, graph, graphviz.Format(graphviz.DOT), f)
	if err != nil {
		slog.Error("Error rendering graph", "error", err)
		os.Exit(1)
	}
}
