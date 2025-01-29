// error.go

package repomap

import (
	"errors"
	"fmt"
)

type RepoMapError struct {
	Err error
}

func (e *RepoMapError) Error() string {
	return e.Err.Error()
}

func (e *RepoMapError) Unwrap() error {
	return e.Err
}

func NewIoError() *RepoMapError {
	return &RepoMapError{Err: errors.New("i/o error")}
}

func NewParseError(msg string) *RepoMapError {
	return &RepoMapError{Err: fmt.Errorf("parse error: %s", msg)}
}

func NewSymbolAnalysisError(msg string) *RepoMapError {
	return &RepoMapError{Err: fmt.Errorf("symbol analysis error: %s", msg)}
}

func NewGraphAnalysisError(msg string) *RepoMapError {
	return &RepoMapError{Err: fmt.Errorf("graph analysis error: %s", msg)}
}

func NewTreeGenerationError(msg string) *RepoMapError {
	return &RepoMapError{Err: fmt.Errorf("tree generation error: %s", msg)}
}

func NewFileSystemError(err error) *RepoMapError {
	return &RepoMapError{Err: fmt.Errorf("filesystem error: %w", err)}
}
