package services

import "fmt"

// xoauth2Client implements the XOAUTH2 SASL mechanism for IMAP.
// Format: user=<user>\x01auth=Bearer <token>\x01\x01
type xoauth2Client struct {
	username string
	token    string
}

func (c *xoauth2Client) Start() (mech string, ir []byte, err error) {
	mech = "XOAUTH2"
	ir = []byte(fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", c.username, c.token))
	return
}

func (c *xoauth2Client) Next(challenge []byte) ([]byte, error) {
	return nil, fmt.Errorf("unexpected challenge: %s", string(challenge))
}
