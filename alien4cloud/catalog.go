package alien4cloud

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// CatalogService is the interface to the service mamaging a4c catalog
type CatalogService interface {
	// UploadCSAR submits a Cloud Service ARchive to Alien4Cloud catalog
	//
	// The csar should be a zip archive containing a single YAML TOSCA definition file at the root of the archive.
	// CSAR could be uploaded into a given workspace, this is a premium feature leave empty on OSS version.
	// If workspace is empty the default workspace will be used.
	//
	// A critical note is that this function may return a ParsingErr. ParsingErr may contain only warnings
	// or informative errors that could be ignored. This can be checked by type casting into a ParsingErr
	// and calling HasCriticalErrors() function.
	UploadCSAR(ctx context.Context, csar io.Reader, workspace string) (csarDefinition CSAR, err error)
}

type catalogService struct {
	client restClient
}

// ParsingErr is an error returned in case of parsing error
// Those parsing errors could be critical or just informative
// HasCriticalErrors() allows to know if this error could be ignored
type ParsingErr interface {
	error
	HasCriticalErrors() bool
	ParsingErrors() map[string][]ParsingError
}

type parsingErr struct {
	parsingErrors map[string][]ParsingError
}

func (pe *parsingErr) Error() string {
	var b strings.Builder
	first := true
	for fileName, errors := range pe.parsingErrors {
		for _, pe := range errors {
			if !first {
				b.WriteString("\n")
			}
			b.WriteString(fileName)
			b.WriteString("> ")
			b.WriteString(pe.String())
			first = false
		}
	}
	return b.String()
}

func (pe *parsingErr) HasCriticalErrors() bool {
	for _, errors := range pe.parsingErrors {
		for _, pe := range errors {
			if pe.ErrorLevel == "ERROR" {
				return true
			}
		}
	}
	return false
}

func (pe *parsingErr) ParsingErrors() map[string][]ParsingError {
	return pe.parsingErrors
}

func (cs *catalogService) UploadCSAR(ctx context.Context, csar io.Reader, workspace string) (CSAR, error) {
	c := CSAR{}
	u := fmt.Sprintf("%s/csars", a4CRestAPIPrefix)
	if workspace != "" {
		u += "?workspace=" + url.QueryEscape(workspace)
	}

	var b bytes.Buffer
	m := multipart.NewWriter(&b)
	if x, ok := csar.(io.Closer); ok {
		defer x.Close()
	}
	fw, err := m.CreateFormFile("file", "types.zip")
	if err != nil {
		return c, errors.Wrap(err, "Cannot create multipart request")
	}
	_, err = io.Copy(fw, csar)
	if err != nil {
		return c, errors.Wrap(err, "Cannot copy multipart request data")
	}
	m.Close()

	response, err := cs.client.doWithContext(ctx, "POST", u, b.Bytes(), []Header{{"Content-Type", m.FormDataContentType()}})
	if err != nil {
		return c, errors.Wrap(err, "Cannot send a request in order to upload a CSAR")
	}

	var res struct {
		Data struct {
			CSAR   CSAR                      `json:"csar,omitempty"`
			Errors map[string][]ParsingError `json:"errors,omitempty"`
		} `json:"data"`
	}

	err = processA4CResponse(response, &res, http.StatusOK)
	if err != nil {
		return c, errors.Wrap(err, "Cannot convert the body of the uploaded CSAR description")
	}

	if len(res.Data.Errors) > 0 {
		err = &parsingErr{res.Data.Errors}
	}
	return res.Data.CSAR, err
}
