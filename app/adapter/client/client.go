package client

import "github.com/kakkky/kakkky.dev/domain"

type Client struct {
	imageUploader *S3ImageUploader
}

func NewClient(imageUploader *S3ImageUploader) *Client {
	return &Client{
		imageUploader: imageUploader,
	}
}

func (c *Client) NewImageUploader() domain.ImageUploader {
	return c.imageUploader
}
