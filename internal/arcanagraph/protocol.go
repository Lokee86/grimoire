package arcanagraph

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

const protocolID = "arcana.query.v1"

type protocolRequest struct {
	ID              string   `json:"id"`
	Op              string   `json:"op"`
	Name            string   `json:"name,omitempty"`
	Kind            string   `json:"kind,omitempty"`
	Path            string   `json:"path,omitempty"`
	Limit           int      `json:"limit,omitempty"`
	NodeID          *uint32  `json:"node_id,omitempty"`
	Direction       string   `json:"direction,omitempty"`
	FromNodeID      *uint32  `json:"from_node_id,omitempty"`
	ToNodeID        *uint32  `json:"to_node_id,omitempty"`
	Relations       []string `json:"relations,omitempty"`
	IncludePossible *bool    `json:"include_possible,omitempty"`
	MaxDepth        int      `json:"max_depth,omitempty"`
}

type protocolResponse struct {
	Protocol string          `json:"protocol"`
	ID       string          `json:"id"`
	OK       bool            `json:"ok"`
	Result   json.RawMessage `json:"result"`
	Error    *protocolError  `json:"error,omitempty"`
}

type protocolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type protocolRun func(context.Context, string, string, []protocolRequest) (map[string]protocolResponse, error)

func runProtocol(
	ctx context.Context,
	command string,
	snapshot string,
	requests []protocolRequest,
) (map[string]protocolResponse, error) {
	var input bytes.Buffer
	encoder := json.NewEncoder(&input)
	for _, request := range requests {
		if err := encoder.Encode(request); err != nil {
			return nil, fmt.Errorf("encode Arcana request: %w", err)
		}
	}

	var stdout, stderr bytes.Buffer
	process := exec.CommandContext(ctx, command, "protocol", "--snapshot", snapshot)
	process.Stdin = &input
	process.Stdout = &stdout
	process.Stderr = &stderr
	if err := process.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return nil, fmt.Errorf("run Arcana protocol: %w: %s", err, message)
		}
		return nil, fmt.Errorf("run Arcana protocol: %w", err)
	}

	responses := make(map[string]protocolResponse, len(requests))
	scanner := bufio.NewScanner(&stdout)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		var response protocolResponse
		if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
			return nil, fmt.Errorf("decode Arcana response: %w", err)
		}
		if response.Protocol != protocolID {
			return nil, fmt.Errorf("unexpected Arcana protocol %q", response.Protocol)
		}
		responses[response.ID] = response
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read Arcana responses: %w", err)
	}
	if err := validateResponses(requests, responses); err != nil {
		return nil, err
	}
	return responses, nil
}

func validateResponses(
	requests []protocolRequest,
	responses map[string]protocolResponse,
) error {
	if len(responses) != len(requests) {
		return fmt.Errorf("Arcana returned %d responses for %d requests", len(responses), len(requests))
	}
	for _, request := range requests {
		response, exists := responses[request.ID]
		if !exists {
			return fmt.Errorf("Arcana did not return response %q", request.ID)
		}
		if response.OK {
			continue
		}
		if response.Error == nil {
			return fmt.Errorf("Arcana operation %s failed without an error payload", request.Op)
		}
		return fmt.Errorf(
			"Arcana operation %s failed (%s): %s",
			request.Op, response.Error.Code, response.Error.Message,
		)
	}
	return nil
}
