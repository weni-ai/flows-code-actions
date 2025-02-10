package codelib

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LanguageType string

const (
	TypePy LanguageType = "python"
)

type CodeLib struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string             `bson:"name,omitempty" json:"name,omitempty"`
	Language LanguageType       `bson:"language,omitempty" json:"language,omitempty"`

	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type UseCase interface {
	Create(ctx context.Context, codelib *CodeLib) (*CodeLib, error)
	CreateBulk(ctx context.Context, codelibs []*CodeLib) ([]*CodeLib, error)
	List(ctx context.Context, Language *LanguageType) ([]CodeLib, error)
	Find(ctx context.Context, name string, language *LanguageType) (*CodeLib, error)
}

type Repository interface {
	Create(context.Context, *CodeLib) (*CodeLib, error)
	CreateBulk(ctx context.Context, codelibs []*CodeLib) ([]*CodeLib, error)
	List(ctx context.Context, Language *LanguageType) ([]CodeLib, error)
	Find(ctx context.Context, name string, language *LanguageType) (*CodeLib, error)
}

func NewCodeLib(name string, language LanguageType) *CodeLib {
	return &CodeLib{
		Name:     name,
		Language: language,
	}
}

func ExtractPythonLibs(pythonCode string) []string {
	standardLibraries := []string{"base64", "datetime", "email", "hashlib", "imaplib", "io", "json", "math", "os", "random", "re", "sys", "tempfile", "time", "urllib", "urllib.parse", "urllib.request", "wave"} // must be alphabetically ordered
	re := regexp.MustCompile(`^(from|import)\s+([\w.]+)`)

	var libraries []string
	for _, line := range strings.Split(pythonCode, "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) > 2 {
			library := strings.Split(matches[2], ".")[0]

			if !contains(standardLibraries, library) {
				libraries = append(libraries, library)
			}
		}
	}

	return removeDoubles(libraries)
}

func InstallPythonLibs(libs []string) error {
	log.Println("Installing python libs")
	for _, lib := range libs {
		// cmd := exec.Command("pip", "install", lib)
		cmd := exec.Command("pip", "install", "--no-cache-dir", lib)
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if stderr.String() != "" {
			log.Println("install lib stderr: ", stderr.String())
		}
		if stdout.String() != "" {
			log.Println("install lib stdout: ", stdout.String())
		}
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error on install lib: %s", lib))
		}
		log.Printf("lib installed: %s\n", lib)
	}
	return nil
}

func contains(s []string, e string) bool {
	i := sort.SearchStrings(s, e)
	return i < len(s) && s[i] == e
}

func removeDoubles(s []string) []string {
	sort.Strings(s)
	j := 0
	for i := 1; i < len(s); i++ {
		if s[j] != s[i] {
			j++
			s[j] = s[i]
		}
	}
	if len(s) > 0 {
		return s[:j+1]
	}
	return s
}
