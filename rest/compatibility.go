package rest

import "context"

// GetOpenApi preserves the v0.1.0 spelling. New code can use GetOpenAPI.
func (c *Client) GetOpenApi(ctx context.Context, options *Options) (*Response, error) {
	return c.GetOpenAPI(ctx, options)
}
