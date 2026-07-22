package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	Endpoint   string
	HTTPClient *http.Client
}

type ProbeResult struct {
	Dimensions int     `json:"dimensions"`
	Similarity float64 `json:"similarity"`
}

type embeddingsRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format"`
}

type embeddingsResponse struct {
	Data []struct {
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewClient(endpoint string) *Client {
	if strings.TrimSpace(endpoint) == "" {
		endpoint = DefaultEndpoint
	}
	return &Client{
		Endpoint: strings.TrimRight(endpoint, "/"),
		HTTPClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

func (client *Client) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	vectors, err := client.EmbedQueries(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	return vectors[0], nil
}

func (client *Client) EmbedDocuments(ctx context.Context, documents []string) ([][]float32, error) {
	if len(documents) == 0 {
		return nil, errors.New("no embedding documents supplied")
	}
	for _, document := range documents {
		if strings.TrimSpace(document) == "" {
			return nil, errors.New("embedding document is empty")
		}
	}
	return client.embed(ctx, documents)
}

func (client *Client) Probe(ctx context.Context, query, document string) (ProbeResult, error) {
	queryVector, err := client.EmbedQuery(ctx, query)
	if err != nil {
		return ProbeResult{}, err
	}
	documentVectors, err := client.EmbedDocuments(ctx, []string{document})
	if err != nil {
		return ProbeResult{}, err
	}
	return ProbeResult{
		Dimensions: len(queryVector),
		Similarity: Dot(queryVector, documentVectors[0]),
	}, nil
}

func (client *Client) embed(ctx context.Context, input []string) ([][]float32, error) {
	body, err := json.Marshal(embeddingsRequest{
		Input: input, Model: ModelReference, EncodingFormat: "float",
	})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.embeddingsURL(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.httpClient().Do(request)
	if err != nil {
		return nil, fmt.Errorf("request embeddings: %w", err)
	}
	defer response.Body.Close()

	var decoded embeddingsResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode embeddings response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message := response.Status
		if decoded.Error != nil && decoded.Error.Message != "" {
			message = decoded.Error.Message
		}
		return nil, fmt.Errorf("embedding service returned %s: %s", response.Status, message)
	}
	if len(decoded.Data) != len(input) {
		return nil, fmt.Errorf("embedding service returned %d vectors for %d inputs", len(decoded.Data), len(input))
	}

	vectors := make([][]float32, len(input))
	for _, item := range decoded.Data {
		if item.Index < 0 || item.Index >= len(vectors) {
			return nil, fmt.Errorf("embedding service returned invalid index %d", item.Index)
		}
		if vectors[item.Index] != nil {
			return nil, fmt.Errorf("embedding service returned duplicate index %d", item.Index)
		}
		vector, err := normalizeTruncated(item.Embedding, Dimensions)
		if err != nil {
			return nil, fmt.Errorf("embedding %d: %w", item.Index, err)
		}
		vectors[item.Index] = vector
	}
	return vectors, nil
}

func (client *Client) embeddingsURL() string {
	if strings.HasSuffix(client.Endpoint, "/embeddings") {
		return client.Endpoint
	}
	return client.Endpoint + "/embeddings"
}

func (client *Client) httpClient() *http.Client {
	if client.HTTPClient != nil {
		return client.HTTPClient
	}
	return http.DefaultClient
}

func normalizeTruncated(input []float64, dimensions int) ([]float32, error) {
	if len(input) < dimensions {
		return nil, fmt.Errorf("received %d dimensions, need at least %d", len(input), dimensions)
	}

	normSquared := 0.0
	for _, value := range input[:dimensions] {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, errors.New("vector contains a non-finite value")
		}
		normSquared += value * value
	}
	if normSquared == 0 {
		return nil, errors.New("vector has zero norm")
	}

	norm := math.Sqrt(normSquared)
	result := make([]float32, dimensions)
	for index, value := range input[:dimensions] {
		result[index] = float32(value / norm)
	}
	return result, nil
}

func Dot(left, right []float32) float64 {
	if len(left) != len(right) {
		return 0
	}
	result := 0.0
	for index := range left {
		result += float64(left[index]) * float64(right[index])
	}
	return result
}
