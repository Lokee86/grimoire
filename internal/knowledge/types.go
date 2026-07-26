package knowledge

import (
	"context"
	"time"
)

const FormatVersion = 1

type Kind string

const (
	KindMarkdown     Kind = "markdown"
	KindText         Kind = "text"
	KindArchitecture Kind = "architecture"
	KindPlanning     Kind = "planning"
	KindADR          Kind = "adr"
	KindIssue        Kind = "issue"
	KindExport       Kind = "export"
	KindConfig       Kind = "config"
)

type Section struct {
	ID          string          `json:"id"`
	Heading     string          `json:"heading"`
	HeadingPath []string        `json:"heading_path,omitempty"`
	StartByte   int             `json:"start_byte"`
	EndByte     int             `json:"end_byte"`
	StartLine   int             `json:"start_line"`
	EndLine     int             `json:"end_line"`
	Hash        string          `json:"hash"`
	Text        string          `json:"text"`
	Terms       []TermFrequency `json:"terms"`
	CodeLinks   []CodeLink      `json:"code_links,omitempty"`
}

type TermFrequency struct {
	Term      string `json:"term"`
	Frequency int    `json:"frequency"`
}

type Document struct {
	ID         string     `json:"id"`
	Path       string     `json:"path"`
	Kind       Kind       `json:"kind"`
	Hash       string     `json:"hash"`
	Size       int64      `json:"size"`
	CommitID   string     `json:"commit_id,omitempty"`
	CommitTime *time.Time `json:"commit_time,omitempty"`
	Sections   []Section  `json:"sections"`
}

type Index struct {
	Version           int        `json:"version"`
	Root              string     `json:"root"`
	GitCommit         string     `json:"git_commit,omitempty"`
	GitTime           *time.Time `json:"git_time,omitempty"`
	SourceFingerprint string     `json:"source_fingerprint,omitempty"`
	Documents         []Document `json:"documents"`
}

type BuildOptions struct {
	IgnoreFile    string
	ExcludePaths  []string
	MaxFileBytes  int64
	IncludeConfig bool
}

type BuildStats struct {
	Scanned int `json:"scanned"`
	Reused  int `json:"reused"`
	Updated int `json:"updated"`
	Removed int `json:"removed"`
}

type SearchOptions struct {
	TopK     int
	Path     string
	Kind     Kind
	Heading  string
	CommitID string
	Since    time.Time
	Until    time.Time
	Vector   VectorRanker
}

type VectorRanker interface {
	Rank(context.Context, string, []Section) (map[string]float64, error)
}

type CodeLink struct {
	Kind       string `json:"kind"`
	Value      string `json:"value"`
	SourcePath string `json:"source_path"`
	Evidence   string `json:"evidence"`
}

type Result struct {
	Handle      string     `json:"handle"`
	DocumentID  string     `json:"document_id"`
	SectionID   string     `json:"section_id"`
	Path        string     `json:"path"`
	Kind        Kind       `json:"kind"`
	Heading     string     `json:"heading"`
	HeadingPath []string   `json:"heading_path,omitempty"`
	StartByte   int        `json:"start_byte"`
	EndByte     int        `json:"end_byte"`
	StartLine   int        `json:"start_line"`
	EndLine     int        `json:"end_line"`
	Hash        string     `json:"hash"`
	Text        string     `json:"text"`
	CommitID    string     `json:"commit_id,omitempty"`
	CommitTime  *time.Time `json:"commit_time,omitempty"`
	Score       float64    `json:"score"`
	Reasons     []string   `json:"reasons"`
	CodeLinks   []CodeLink `json:"code_links,omitempty"`
}

type SearchResponse struct {
	Results     []Result `json:"results"`
	VectorUsed  bool     `json:"vector_used"`
	VectorError string   `json:"vector_error,omitempty"`
}

func (document Document) Handle(section Section) string {
	return "knowledge://" + document.Path + "#" + section.ID
}

func (document Document) Section(sectionID string) (Section, bool) {
	for _, section := range document.Sections {
		if section.ID == sectionID {
			return section, true
		}
	}
	return Section{}, false
}
