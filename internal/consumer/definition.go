package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

func (d *Definition) UnmarshalJSON(data []byte) error {
	var value struct {
		Version int             `json:"version"`
		Command string          `json:"command"`
		Args    []string        `json:"args"`
		Timeout json.RawMessage `json:"timeout"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	timeout, err := decodeTimeout(value.Timeout)
	if err != nil {
		return err
	}
	*d = Definition{
		Version: value.Version,
		Command: value.Command,
		Args:    value.Args,
		Timeout: timeout,
	}
	return nil
}

func (d Definition) MarshalJSON() ([]byte, error) {
	timeout := ""
	if d.Timeout != 0 {
		timeout = d.Timeout.String()
	}
	return json.Marshal(struct {
		Version int      `json:"version"`
		Command string   `json:"command"`
		Args    []string `json:"args,omitempty"`
		Timeout string   `json:"timeout,omitempty"`
	}{d.Version, d.Command, d.Args, timeout})
}

func decodeTimeout(data json.RawMessage) (time.Duration, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return 0, nil
	}
	var text string
	if data[0] == '"' {
		if err := json.Unmarshal(data, &text); err != nil {
			return 0, err
		}
		duration, err := time.ParseDuration(text)
		if err != nil {
			return 0, fmt.Errorf("invalid consumer timeout %q: %w", text, err)
		}
		if duration < 0 {
			return 0, fmt.Errorf("consumer timeout must not be negative")
		}
		return duration, nil
	}
	var nanos int64
	if err := json.Unmarshal(data, &nanos); err != nil {
		return 0, fmt.Errorf("consumer timeout must be a duration string: %w", err)
	}
	if nanos < 0 {
		return 0, fmt.Errorf("consumer timeout must not be negative")
	}
	return time.Duration(nanos), nil
}
