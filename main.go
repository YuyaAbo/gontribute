package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v45/github"
	"golang.org/x/mod/modfile"
	"golang.org/x/oauth2"
)

var (
	githubToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	owner       = os.Getenv("GITHUB_TARGET_OWNER")
	repo        = os.Getenv("GITHUB_TARGET_REPO")
)

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	fc, _, _, err := client.Repositories.GetContents(ctx, owner, repo, "go.mod", nil)
	if err != nil {
		log.Fatalln(err)
	}

	s, err := fc.GetContent()
	if err != nil {
		log.Fatalln(err)
	}

	f, err := modfile.Parse("go.mod", []byte(s), nil)
	if err != nil {
		log.Fatalln(err)
	}

	for _, req := range f.Require {
		if req.Indirect {
			continue
		}
		if !strings.HasPrefix(req.Mod.Path, "github.com/") {
			continue
		}

		func() {
			path := strings.TrimPrefix(req.Mod.Path, "github.com/")
			reqOwner := strings.Split(path, "/")[0]
			reqRepo := strings.Split(path, "/")[1]
			url, _, err := client.Repositories.GetArchiveLink(ctx, reqOwner, reqRepo, github.Zipball, nil, true)
			if err != nil {
				log.Println(err)
				return
			}

			resp, err := http.Get(url.String())
			if err != nil {
				log.Println(err)
				return
			}
			defer resp.Body.Close()

			zipFileName := fmt.Sprintf("%s-%s.zip", reqOwner, reqRepo)
			out, err := os.Create(zipFileName)
			if err != nil {
				log.Println(err)
				return
			}
			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				log.Println(err)
				return
			}

			b, err := exec.Command("zipinfo", "-1", zipFileName).Output()
			if err != nil {
				log.Println(err)
				return
			}
			dirName := strings.Split(string(b), "/")[0]

			if err = exec.Command("unzip", "-o", zipFileName).Run(); err != nil {
				log.Println(err)
				return
			}

			if err = exec.Command("cd", dirName).Run(); err != nil {
				log.Println(err)
				return
			}

			fmt.Printf("## %s/%s\n", reqOwner, reqRepo)

			b, err = exec.Command("goreportcard-cli").CombinedOutput()
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Println(string(b))

			if err = exec.Command("cd", "..").Run(); err != nil {
				log.Println(err)
				return
			}

			if err = exec.Command("rm", "-rf", dirName).Run(); err != nil {
				log.Println(err)
				return
			}

			if err = exec.Command("rm", zipFileName).Run(); err != nil {
				log.Println(err)
				return
			}
		}()
	}
}
