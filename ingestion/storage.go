package ingestion

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v48/github"
	"io"
	"os"
	"path/filepath"
)

type ArtifactStore struct {
	RootPath string
}

func (as ArtifactStore) PathToArtifact(artifact *github.Artifact) string {
	zipfileName := fmt.Sprintf("%s.zip", *artifact.Name)
	return filepath.Join(as.RootPath, fmt.Sprintf("%d", *artifact.WorkflowRunMetadata.ID), zipfileName)
}

func (as ArtifactStore) ArtifactExists(artifact *github.Artifact) bool {
	_, err := os.Stat(as.PathToArtifact(artifact))
	return !os.IsNotExist(err)
}

func (as ArtifactStore) ListArtifacts() ([]*github.Artifact, error) {
	if metadataFiles, err := filepath.Glob(filepath.Join(as.RootPath, "**", "*.json")); err != nil {
		return nil, fmt.Errorf("failed to list metadata files: %w", err)
	} else {
		artifacts := make([]*github.Artifact, 0)

		for _, metadataFile := range metadataFiles {
			fmt.Printf("Loading metadata file %s\n", metadataFile)
			if artifact, err := loadArtifact(metadataFile); err != nil {
				return nil, fmt.Errorf("failed to load artifact from %s: %w", metadataFile, err)
			} else {
				fmt.Printf("Loaded artifact %s\n", *artifact.Name)
				artifacts = append(artifacts, artifact)
			}
		}

		return artifacts, nil
	}
}

func (as ArtifactStore) Store(artifact *github.Artifact, filedata []byte) error {
	workflowDir := filepath.Join(as.RootPath, fmt.Sprintf("%d", *artifact.WorkflowRunMetadata.ID))
	zipfileName := fmt.Sprintf("%s.zip", *artifact.Name)
	metadataFileName := fmt.Sprintf("%s.json", *artifact.Name)

	artifactPath := filepath.Join(workflowDir, zipfileName)
	metadataFilePath := filepath.Join(workflowDir, metadataFileName)

	fmt.Printf("Storing artifact %s to %s\n", *artifact.Name, artifactPath)

	if err := os.MkdirAll(workflowDir, os.ModePerm); err != nil {
		return err
	}

	if err := writeFile(artifactPath, filedata); err != nil {
		return fmt.Errorf("failed to write artifact file %s: %w", artifactPath, err)
	} else {
		if b, err := json.Marshal(artifact); err != nil {
			return fmt.Errorf("failed to marshal artifact metadata: %w", err)
		} else {
			if err := writeFile(metadataFilePath, b); err != nil {
				return fmt.Errorf("failed to write metadata file %s: %w", metadataFilePath, err)
			}
		}
	}

	return nil
}

func loadArtifact(path string) (*github.Artifact, error) {

	if f, err := os.Open(path); err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	} else {
		if b, err := io.ReadAll(f); err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", path, err)
		} else {
			artifact := &github.Artifact{}
			if err := json.Unmarshal(b, artifact); err != nil {
				return nil, fmt.Errorf("failed to unmarshal artifact: %w", err)
			} else {
				return artifact, nil
			}
		}
	}
}

func writeFile(path string, data []byte) error {
	if f, err := os.Create(path); err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	} else {
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

		if _, err := f.Write(data); err != nil {
			return err
		}
	}

	return nil
}
