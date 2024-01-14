package bunnystorage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
	"git.sr.ht/~jamesponddotco/xstd-go/xstrings"
	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
)

const (
	// ErrConfigRequired is returned when a Client is created without a Config.
	ErrConfigRequired xerrors.Error = "config is required"
)

type (
	// Client is the LanguageTool API client.
	Client struct {
		// httpc is the underlying HTTP client used by the API client.
		httpc *resty.Client

		// cfg specifies the configuration used by the API client.
		cfg *Config
	}
)

// NewClient returns a new bunny.net Edge Storage API client.
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, ErrConfigRequired
	}

	cfg.init()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	c := &Client{
		httpc: resty.New().AddRetryAfterErrorCondition().SetRetryCount(5).SetRateLimiter(rate.NewLimiter(rate.Every(time.Second/50), 100)),

		cfg: cfg,
	}

	if cfg.Debug {
		c.httpc = c.httpc.EnableTrace()
	}

	return c, nil
}

// List lists the files in the storage zone.
func (c *Client) List(ctx context.Context, path string) ([]*Object, *resty.Response, error) {
	path = strings.TrimPrefix(path, "/")

	uri := xstrings.JoinWithSeparator("/", c.cfg.Endpoint.String(), c.cfg.StorageZone, path+"/")

	headers := map[string][]string{
		"Accept":    {"application/json"},
		"AccessKey": {c.cfg.AccessKey(OperationRead)},
	}

	resp, err := c.httpc.R().SetHeaderMultiValues(headers).Get(uri)
	if err != nil {
		return nil, resp, fmt.Errorf("%w", err)
	}

	var files []*Object
	if err := json.Unmarshal(resp.Body(), &files); err != nil {
		return nil, resp, fmt.Errorf("%w", err)
	}

	return files, resp, nil
}

// Download downloads a file from the storage zone.
func (c *Client) Download(ctx context.Context, path, filename string) ([]byte, *resty.Response, error) {
	path = strings.TrimPrefix(path, "/")
	filename = filepath.Base(filename)

	uri := xstrings.JoinWithSeparator("/", c.cfg.Endpoint.String(), c.cfg.StorageZone, path, filename)

	headers := map[string][]string{
		"Accept":    {"*/*"},
		"AccessKey": {c.cfg.AccessKey(OperationRead)},
	}

	resp, err := c.httpc.R().SetHeaderMultiValues(headers).Get(uri)
	if err != nil {
		return nil, resp, fmt.Errorf("%w", err)
	}

	return resp.Body(), resp, nil
}

// Upload uploads a file to the storage zone.
func (c *Client) Upload(ctx context.Context, path, filename, checksum string, body io.Reader) (*resty.Response, error) {
	path = strings.TrimPrefix(path, "/")

	uri := xstrings.JoinWithSeparator("/", c.cfg.Endpoint.String(), c.cfg.StorageZone, path, filename)

	headers := map[string][]string{
		"AccessKey": {c.cfg.AccessKey(OperationWrite)},
	}

	if checksum != "" {
		headers["Checksum"] = []string{strings.ToUpper(checksum)}
	}

	resp, err := c.httpc.R().SetHeaderMultiValues(headers).SetBody(body).Put(uri)
	if err != nil {
		return resp, fmt.Errorf("%w", err)
	}

	return resp, nil
}

// Delete deletes a file from the storage zone.
func (c *Client) Delete(ctx context.Context, path, filename string) (*resty.Response, error) {
	path = strings.TrimPrefix(path, "/")
	filename = filepath.Base(filename)

	uri := xstrings.JoinWithSeparator("/", c.cfg.Endpoint.String(), c.cfg.StorageZone, path, filename)

	headers := map[string][]string{
		"AccessKey": {c.cfg.AccessKey(OperationWrite)},
	}

	resp, err := c.httpc.R().SetHeaderMultiValues(headers).Delete(uri)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return resp, nil
}
