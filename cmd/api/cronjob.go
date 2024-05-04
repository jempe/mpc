package main

import (
	"fmt"
	"strconv"

	"github.com/jempe/mpc/internal/data"
	"github.com/pgvector/pgvector-go"
)

func (app *application) runCronJob() {
	app.logger.PrintInfo("starting the cron job", nil)

	var embeddingsProviders []string

	if app.config.sentenceTransformersEnable {
		embeddingsProviders = append(embeddingsProviders, "sentence-transformers")
	}

	if app.config.openAIApiKey != "" {
		embeddingsProviders = append(embeddingsProviders, "openai")
	}

	maxTokens := 200

	itemsPerBatch := app.config.embeddingsPerBatch

	if len(embeddingsProviders) > 0 {

		filter := data.Filters{
			Page:     1,
			PageSize: itemsPerBatch,
			Sort:     "-id",
			SortSafelist: []string{
				"-id",
			},
		}

		videos, metadata, err := app.models.Videos.GetAllNotInSemantic(filter)
		if err != nil {
			app.logger.PrintError(err, nil)
			return
		}

		app.logger.PrintInfo("Videos to process", map[string]string{
			"total":      strconv.Itoa(metadata.TotalRecords),
			"processing": strconv.Itoa(len(videos)),
		})

		for _, video := range videos {

			fields := []string{
				"videos.description",
			}

			for _, field := range fields {
				content := ""

				switch field {
				case "videos.description":
					content = video.Description
				}

				if content != "" {
					countedTokens, err := app.countTokens(content)
					if err != nil {
						app.logger.PrintError(err, nil)
						return
					}

					if countedTokens < maxTokens {

						Document := &data.Document{
							Content:      content,
							Tokens:       countedTokens,
							Sequence:     1,
							VideoID:      video.ID,
							ContentField: field,
						}

						err := app.models.Documents.Insert(Document)
						if err != nil {
							app.logger.PrintError(err, nil)
							return
						}
					} else {
						parts, partsErr := app.splitText(content, maxTokens)
						if partsErr != nil {
							app.logger.PrintError(partsErr, nil)
							return
						}

						for seq, part := range parts {

							countedTokens, err := app.countTokens(part)
							if err != nil {
								app.logger.PrintError(err, nil)
								return
							}

							DocumentPart := &data.Document{
								Content:      part,
								Tokens:       countedTokens,
								Sequence:     seq + 1,
								VideoID:      video.ID,
								ContentField: field,
							}

							err = app.models.Documents.Insert(DocumentPart)
							if err != nil {
								app.logger.PrintError(err, nil)
								return
							}
						}
					}
				}
			}
		}

	}

	for _, provider := range embeddingsProviders {
		Documents, metadata, err := app.models.Documents.GetAllWithoutEmbeddings(itemsPerBatch, provider)
		if err != nil {
			app.logger.PrintError(err, nil)
			return
		}

		app.logger.PrintInfo("Documents to process", map[string]string{
			"total":      strconv.Itoa(metadata.TotalRecords),
			"processing": strconv.Itoa(len(Documents)),
		})

		if len(Documents) == 0 {
			continue
		}

		var embeddingsContent []string

		for _, Document := range Documents {
			embeddingsContent = append(embeddingsContent, Document.Content)
		}

		var embeddings [][]float32

		if provider == "sentence-transformers" {
			embeddings, err = app.fetchSentenceTransformersEmbeddings(embeddingsContent)

			if err != nil {
				app.logger.PrintError(err, nil)
				return
			}
		} else if provider == "openai" {
			embeddings, err = fetchOpenaiEmbeddings(embeddingsContent, app.config.openAIApiKey)

			if err != nil {
				app.logger.PrintError(err, nil)
				return
			}
		}

		app.logger.PrintInfo(fmt.Sprintf("Embeddings fetched from %s", provider), map[string]string{
			"total": strconv.Itoa(len(embeddings)),
		})

		for i, Document := range Documents {
			app.models.Documents.UpdateEmbedding(Document, pgvector.NewVector(embeddings[i]), provider)
		}

	}
}
