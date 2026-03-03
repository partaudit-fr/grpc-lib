package files

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/metadata"
)

const (
	defaultBufSize = 1 << 20  // 1 MB
	maxMemory      = 32 << 20 // 32 MB. parameter for ReadForm.
)

var ErrSizeLimitExceeded = errors.New("size limit exceeded")

func copyZeroAlloc(w io.Writer, r io.Reader) (int64, error) {
	if wt, ok := r.(io.WriterTo); ok {
		return wt.WriteTo(w)
	}

	if rt, ok := w.(io.ReaderFrom); ok {
		return rt.ReadFrom(r)
	}

	vbuf := copyBufPool.Get()
	buf := vbuf.([]byte)

	n, err := io.CopyBuffer(w, r, buf)

	copyBufPool.Put(vbuf)

	return n, err
}

var copyBufPool = sync.Pool{
	New: func() any {
		return make([]byte, 4096)
	},
}

// SaveMultipartFile saves multipart file fh under the given filename path.
func SaveMultipartFile(fh *multipart.FileHeader, path string) (err error) {
	var (
		f  multipart.File
		ff *os.File
	)

	f, err = fh.Open()
	if err != nil {
		return
	}

	var ok bool
	if ff, ok = f.(*os.File); ok {
		// Windows can't rename files that are opened.
		if err = f.Close(); err != nil {
			return
		}

		// If renaming fails we try the normal copying method.
		// Renaming could fail if the files are on different devices.
		if os.Rename(ff.Name(), path) == nil {
			return nil
		}

		// Reopen f for the code below.
		if f, err = fh.Open(); err != nil {
			return
		}
	}

	defer func() {
		e := f.Close()
		if err == nil {
			err = e
		}
	}()

	if ff, err = os.Create(path); err != nil {
		return
	}

	defer func() {
		e := ff.Close()
		if err == nil {
			err = e
		}
	}()

	_, err = copyZeroAlloc(ff, f)

	return
}

type FormData struct {
	form *multipart.Form
}

func NewFormData(ctx context.Context, body *httpbody.HttpBody) (*FormData, error) {
	boundary, err := extractBoundaryFromContext(ctx)
	if err != nil {
		return nil, err
	}

	reader := multipart.NewReader(bytes.NewReader(body.GetData()), boundary)

	form, err := reader.ReadForm(maxMemory)
	if err != nil {
		return nil, err
	}

	return &FormData{form: form}, err
}

func (f *FormData) Value(key string) []string {
	return f.form.Value[key]
}

func (f *FormData) Files(key string) ([]*multipart.FileHeader, error) {
	files := f.form.File[key]

	for _, file := range files {
		if file.Filename == "" {
			return nil, errors.New("filename is empty")
		}
	}

	return files, nil
}

func (f *FormData) RemoveAll() error {
	return f.form.RemoveAll()
}

// AllValues returns all form value keys for debugging
func (f *FormData) AllValues() map[string][]string {
	return f.form.Value
}

// AllFileKeys returns all form file keys for debugging
func (f *FormData) AllFileKeys() []string {
	keys := make([]string, 0, len(f.form.File))
	for k := range f.form.File {
		keys = append(keys, k)
	}
	return keys
}

// extractBoundaryFromContext retrieves the boundary from the content-type metadata.
func extractBoundaryFromContext(ctx context.Context) (string, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	contentType := md.Get(fmt.Sprintf("%s%s", runtime.MetadataPrefix, "content-type"))

	if len(contentType) == 0 {
		return "", http.ErrNotMultipart
	}

	mediaType, params, err := mime.ParseMediaType(contentType[0])
	if err != nil || !(mediaType == "multipart/form-data" || mediaType == "multipart/mixed") {
		return "", http.ErrNotMultipart
	}

	boundary, ok := params["boundary"]
	if !ok {
		return "", http.ErrMissingBoundary
	}

	return boundary, nil
}
