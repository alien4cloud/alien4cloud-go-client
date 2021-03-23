package alien4cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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

	// GetComplexTOSCAType gets the description of a complex TOSCA type
	GetComplexTOSCAType(ctx context.Context, request ComplexToscaTypeDescriptorRequest) (map[string]interface{}, error)
}

const (
	// TYPE_DESCRIPTION_CONTENT_TYPE_KEY is a key in a map returned by GetComplexTOSCAType()
	// providing details of a property definition of type array or map
	TYPE_DESCRIPTION_CONTENT_TYPE_KEY = "_contentType"
	// TYPE_DESCRIPTION_TYPE_KEY is a key in a map returned by GetComplexTOSCAType()
	// providing the type of a property (complex, tosca, array, map)
	TYPE_DESCRIPTION_TYPE_KEY = "_type"
	// TYPE_DESCRIPTION_PROPERTY_TYPE_KEY is a key in a map returned by GetComplexTOSCAType()
	// providing details on a complex type
	TYPE_DESCRIPTION_PROPERTY_TYPE_KEY = "_propertyType"
	// TYPE_DESCRIPTION_TOSCA_DEFINITION_KEY is a key in a map returned by GetComplexTOSCAType()
	// providing details on a tosca type
	TYPE_DESCRIPTION_TOSCA_DEFINITION_KEY = "_definition"
	// TYPE_DESCRIPTION_TOSCA_TYPE is the type of a non-complex property in a map returned by GetComplexTOSCAType()
	TYPE_DESCRIPTION_TOSCA_TYPE = "tosca"
	// TYPE_DESCRIPTION_COMPLEX_TYPE is the type of a complex property in a map a map returned by GetComplexTOSCAType()
	TYPE_DESCRIPTION_COMPLEX_TYPE = "complex"
	// TYPE_DESCRIPTION_ARRAY_TYPE is the type of aa array property in a map returned by GetComplexTOSCAType()
	TYPE_DESCRIPTION_ARRAY_TYPE = "array"
	// TYPE_DESCRIPTION_MAP_TYPE is the type of a map property in a map returned by GetComplexTOSCAType()
	TYPE_DESCRIPTION_MAP_TYPE = "map"
)

type catalogService struct {
	client *a4cClient
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

	// TODO(loicalbertin) we may have an issue on large files as it will load the whole file in memory.
	// We should consider using io.Pipe() to create a synchronous in-memory pipe.
	// The tricky part will be to make it work with an expected io.ReadSeeker.
	var b bytes.Buffer
	m := multipart.NewWriter(&b)
	defer m.Close()
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

	request, err := cs.client.NewRequest(ctx, "POST", u, bytes.NewReader(b.Bytes()))
	if err != nil {
		return c, errors.Wrap(err, "Cannot create a request in order to upload a CSAR")
	}
	request.Header.Set("Content-Type", m.FormDataContentType())

	var res struct {
		Data struct {
			CSAR   CSAR                      `json:"csar,omitempty"`
			Errors map[string][]ParsingError `json:"errors,omitempty"`
		} `json:"data"`
	}

	response, err := cs.client.Do(request)
	if err != nil {
		return c, errors.Wrap(err, "Cannot send a request in order to upload a CSAR")
	}

	err = ReadA4CResponse(response, &res)
	if err != nil {
		return c, errors.Wrap(err, "Cannot read response on CSAR upload")
	}

	if len(res.Data.Errors) > 0 {
		err = &parsingErr{res.Data.Errors}
	}
	return res.Data.CSAR, err
}

// GetComplexTOSCAType gets the description of a complex TOSCA type
func (cs *catalogService) GetComplexTOSCAType(ctx context.Context, request ComplexToscaTypeDescriptorRequest) (map[string]interface{}, error) {
	jsonReq, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot marshal complex TOSCA type request %+v", request)
	}

	req, err := cs.client.NewRequest(ctx,
		"POST",
		fmt.Sprintf("%s/formdescriptor/complex-tosca-type", a4CRestAPIPrefix),
		bytes.NewReader(jsonReq))
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot create request to get TOSCA type description %v", request)
	}

	var res map[string]interface{}

	response, err := cs.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot send request to get TOSCA type description %v", request)
	}
	err = ReadA4CResponse(response, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read response for TOSCA type description %s", string(jsonReq))
	}
	return res, err
}
