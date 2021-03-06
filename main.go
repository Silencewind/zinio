package main

import (
	"bytes"
	"context"
	"flag"
	"regexp"
	"strings"

	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
	"github.com/unidoc/unidoc/pdf"
	"golang.org/x/net/context/ctxhttp"

	log "github.com/sirupsen/logrus"
)

var (
	invalidChars = regexp.MustCompile("[^\\p{L}\\p{N}\\s'!&,.@-]")
	whiteSpaces  = regexp.MustCompile("\\s+")
)

func sanitize(name string) string {
	r := strings.NewReplacer("/", " - ", ":", " - ")
	s := r.Replace(name)

	s = invalidChars.ReplaceAllString(s, " ")
	s = whiteSpaces.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	return s
}

func downloadPage(ctx context.Context, url string) (page, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := ctxhttp.Get(ctx, http.DefaultClient, url)

	if err != nil {
		return page{}, err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return page{}, err
	}

	r := bytes.NewReader(b)
	return page{r}, nil
}

func downloadAllPages(ctx context.Context, issue *Issue) ([]page, error) {
	var pages []page

	for i := 0; i < issue.PageCount; i++ {
		url, err := issue.GetURL(i)

		if err != nil {
			return nil, err
		}

		page, err := downloadPage(ctx, url)

		if err != nil {
			return nil, errors.Wrapf(err, "failed to download page %d", i)
		}

		pages = append(pages, page)
	}

	if len(pages) > 0 {
		pages = append(pages[1:], pages[0])
	}

	return pages, nil
}

func downloadAllIssues(ctx context.Context, session *Session, magazines []Magazine) error {
	for _, magazine := range magazines {
		dir := sanitize(magazine.Title)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.Mkdir(dir, 0755); err != nil {
				log.Error(errors.Wrapf(err, "failed to create directory %s", magazine.Title))
				continue
			}
		}

		for _, metadata := range magazine.Issues {
			file := sanitize(metadata.Title)
			path := path.Join(dir, file+".pdf")

			entry := log.WithFields(log.Fields{
				"magazine": magazine.Title,
				"issue":    metadata.Title,
			})

			if _, err := os.Stat(path); err == nil {
				entry.Info("issue already downloaded")
				continue
			}

			err := func() error {
				entry.Info("downloading issue metadata")
				issue, err := session.GetIssue(ctx, magazine.ID, metadata.ID)

				if err != nil {
					return err
				}

				entry.Info("downloading issue")
				pages, err := downloadAllPages(ctx, issue)

				if err != nil {
					return errors.Wrapf(err, "failed to download %s %s", magazine.Title, metadata.Title)
				}

				entry.Info("saving issue")
				err = save(session, pages, issue.Password, path)

				if err != nil {
					return err
				}

				return nil
			}()

			if err != nil {
				log.Error(err)
			}
		}
	}

	return nil
}

func save(session *Session, pages []page, password string, path string) error {
	pdf, err := unlockAndMerge(pages, []byte(password))

	if err != nil {
		return errors.Wrapf(err, "failed to unlock and merge pages for %s", path)
	}

	temp := path + ".part"
	file, err := os.Create(temp)

	if err != nil {
		return errors.Wrapf(err, "failed to create %s", path)
	}

	err = pdf.Write(file)
	cerr := file.Close()

	if err != nil || cerr != nil {
		return errors.Wrapf(err, "failed to save %s", path)
	}

	return errors.Wrapf(os.Rename(temp, path), "failed to save %s", path)
}

func unlockAndMerge(pages []page, password []byte) (*pdf.PdfWriter, error) {
	w := pdf.NewPdfWriter()

	for _, page := range pages {
		r, err := pdf.NewPdfReader(page)

		if err != nil {
			return nil, err
		}

		ok, err := r.Decrypt(password)

		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, errors.Errorf("failed to decrypt pages using password %s", string(password))
		}

		numPages, err := r.GetNumPages()

		if err != nil {
			return nil, err
		}

		for i := 0; i < numPages; i++ {
			page, err := r.GetPageAsPdfPage(i + 1)

			if err != nil {
				return nil, err
			}

			page.Annots = nil

			if err = w.AddPage(page.GetPageAsIndirectObject()); err != nil {
				return nil, err
			}
		}
	}

	return &w, nil
}

func main() {
	var login, password string

	flag.StringVar(&login, "email", "", "Account email")
	flag.StringVar(&password, "password", "", "Account password")

	flag.Parse()

	if login == "" {
		login = os.Getenv("ZINIO_EMAIL")
	}

	if password == "" {
		password = os.Getenv("ZINIO_PASSWORD")
	}

	if login == "" || password == "" {
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()

	log.WithField("user", login).Info("logging in")
	session, err := Login(ctx, login, password)

	if err != nil {
		log.Fatal(err)
	}

	log.Info("downloading list of all magazines")
	magazines, err := session.GetMagazines(ctx)

	if err != nil {
		log.Fatal(err)
	}

	if err = downloadAllIssues(ctx, session, magazines); err != nil {
		log.Fatal(err)
	}
}
