package main

import (
	"fmt"
	"strconv"

	"github.com/jempe/mpc/internal/data"
	"github.com/pgvector/pgvector-go"
)

func (app *application) runCronJob() error {
	app.logger.PrintInfo("starting the cron job", nil)

	var embeddingsProviders []string
	if app.config.embeddings.sentenceTransformersServerURL != "" {
		embeddingsProviders = append(embeddingsProviders, "sentence-transformers")
	}

	maxTokens := 200

	itemsPerBatch := app.config.embeddings.embeddingsPerBatch

	if len(embeddingsProviders) == 0 {
		return fmt.Errorf("No embeddings providers configured", nil)
	}

	filter := data.Filters{
		Page:     1,
		PageSize: itemsPerBatch,
		Sort:     "-id",
		SortSafelist: []string{
			"-id",
		},
	}

	var fields []string

	fields = []string{
		"videos.name",
		"videos.description",
	}

	for _, field := range fields {

		videos, metadata, err := app.models.Videos.GetAllNotInSemantic(filter, field)
		if err != nil {
			return err
		}

		app.logger.PrintInfo("Videos to process", map[string]string{
			"field":      field,
			"total":      strconv.Itoa(metadata.TotalRecords),
			"processing": strconv.Itoa(len(videos)),
		})

		for _, video := range videos {

			content := ""

			switch field {
			case "videos.name":
				content = video.Name
			case "videos.description":
				content = video.Description
			}

			titleField := video.Name

			if content != "" {
				countedTokens, err := app.countTokens(content)
				if err != nil {
					return err
				}

				var parts []string
				var partsErr error

				if countedTokens < maxTokens {
					parts = []string{content}
				} else {
					parts, partsErr = app.splitText(content, maxTokens)
					if partsErr != nil {
						return partsErr
					}
				}

				for seq, part := range parts {

					countedTokens, err := app.countTokens(part)
					if err != nil {
						return err
					}

					documentPart := &data.Document{
						Title:        titleField,
						Content:      part,
						Tokens:       countedTokens,
						Sequence:     seq + 1,
						VideoID:      video.ID,
						ContentField: field,
					}

					if countedTokens > app.config.embeddings.maxTokens {
						return fmt.Errorf("Document of Video %d has too many tokens %d, sequence: %d", video.ID, countedTokens, seq)
					} else {
						err = app.models.Documents.Insert(documentPart)
						if err != nil {
							return err
						}
					}
				}
			} else {
				app.logger.PrintInfo("No content to process", map[string]string{
					"video_id": strconv.Itoa(int(video.ID)),
				})
			}
		}
	}

	fields = []string{
		"categories.name",
	}

	for _, field := range fields {

		categories, metadata, err := app.models.Categories.GetAllNotInSemantic(filter, field)
		if err != nil {
			return err
		}

		app.logger.PrintInfo("Categories to process", map[string]string{
			"field":      field,
			"total":      strconv.Itoa(metadata.TotalRecords),
			"processing": strconv.Itoa(len(categories)),
		})

		for _, category := range categories {

			content := ""

			switch field {
			case "categories.name":
				content = category.Name
			}

			titleField := category.Name

			if content != "" {
				countedTokens, err := app.countTokens(content)
				if err != nil {
					return err
				}

				var parts []string
				var partsErr error

				if countedTokens < maxTokens {
					parts = []string{content}
				} else {
					parts, partsErr = app.splitText(content, maxTokens)
					if partsErr != nil {
						return partsErr
					}
				}

				for seq, part := range parts {

					countedTokens, err := app.countTokens(part)
					if err != nil {
						return err
					}

					documentPart := &data.Document{
						Title:        titleField,
						Content:      part,
						Tokens:       countedTokens,
						Sequence:     seq + 1,
						CategoryID:   category.ID,
						ContentField: field,
					}

					if countedTokens > app.config.embeddings.maxTokens {
						return fmt.Errorf("Document of Category %d has too many tokens %d, sequence: %d", category.ID, countedTokens, seq)
					} else {
						err = app.models.Documents.Insert(documentPart)
						if err != nil {
							return err
						}
					}
				}
			} else {
				app.logger.PrintInfo("No content to process", map[string]string{
					"category_id": strconv.Itoa(int(category.ID)),
				})
			}
		}
	}

	for _, provider := range embeddingsProviders {
		documents, metadata, err := app.models.Documents.GetAllWithoutEmbeddings(itemsPerBatch, provider)
		if err != nil {
			return err
		}

		app.logger.PrintInfo("Documents to process", map[string]string{
			"total":      strconv.Itoa(metadata.TotalRecords),
			"processing": strconv.Itoa(len(documents)),
		})

		if len(documents) == 0 {
			continue
		}

		var embeddingsContent []string

		for _, document := range documents {
			embeddingsContent = append(embeddingsContent, document.Content)
		}

		var embeddings [][]float32
		if provider == "sentence-transformers" {
			embeddings, err = app.fetchSentenceTransformersEmbeddings(embeddingsContent)
		}

		if err != nil {
			return err
		}

		app.logger.PrintInfo(fmt.Sprintf("Embeddings fetched from %s", provider), map[string]string{
			"total": strconv.Itoa(len(embeddings)),
		})

		for i, document := range documents {
			app.models.Documents.UpdateEmbedding(document, pgvector.NewVector(embeddings[i]), provider)
		}
	}
	return nil

}
