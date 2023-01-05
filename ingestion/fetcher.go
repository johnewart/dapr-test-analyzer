package ingestion

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
	"io"
	"log"
	http "net/http"
	"os"
	"path"
	"strings"
)

func FetchResults(limit int, maxPages int) error {
	token := os.Getenv("GITHUB_TOKEN")
	outPath := os.Getenv("CACHE_DIR")
	if outPath == "" {
		if userDir, err := os.UserHomeDir(); err != nil {
			return fmt.Errorf("unable to determine cache dir via environment or default ~/.cache/dapr-test-analyzer")
		} else {
			outPath = path.Join(userDir, ".cache", "dapr-test-analyzer")
		}
	}

	log.Printf("Writing data files to %s", outPath)
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	pages := maxPages
	artifactCount := 0

	store := ArtifactStore{RootPath: outPath}

	for i := 1; i <= pages; i++ {
		if artifactCount >= limit {
			fmt.Printf("Reached limit of %d artifacts\n", limit)
			break
		}

		if artifacts, _, err := client.Actions.ListArtifacts(context.TODO(), "dapr", "dapr", &github.ListOptions{
			Page: i,
		}); err != nil {
			return err
		} else {
			for _, a := range artifacts.Artifacts {
				if strings.Contains(a.GetName(), "e2e") {
					artifactCount += 1
					//fmt.Printf("ID: %d\n", *workflow_run.ID)
					fmt.Printf("*  %s\n", a.GetName())
					if !store.ArtifactExists(a) {
						fmt.Printf("Downloading artifact %s\n", a.GetName())
						artifactUrl, _, err := client.Actions.DownloadArtifact(context.TODO(), "dapr", "dapr", *a.ID, true)
						if err != nil {
							return err
						}

						if artifactData, err := downloadArtifact(artifactUrl.String()); err != nil {
							return err
						} else {
							if err := store.Store(a, artifactData); err != nil {
								return err
							}
						}

					} else {
						fmt.Printf("Artifact %s already exists, skipping\n", a.GetName())
					}
				}
			}
		}
	}

	return nil
}

func downloadArtifact(url string) ([]byte, error) {

	c := http.Client{}
	if req, err := http.NewRequest("GET", url, nil); err != nil {
		return nil, err
	} else {
		if res, err := c.Do(req); err != nil {
			return nil, err
		} else {
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(res.Body)

			return io.ReadAll(res.Body)
		}
	}

}
