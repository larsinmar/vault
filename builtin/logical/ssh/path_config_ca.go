package ssh

import (
	"fmt"
	"github.com/hashicorp/vault/helper/errutil"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"golang.org/x/crypto/ssh"
)

func pathConfigCA(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config/ca",
		Fields: map[string]*framework.FieldSchema{
			"private_key": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: `Private half of the SSH key that will be used to sign certificates.`,
			},
			"public_key": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: `Public half of the SSH key that will be used to sign certificates.`,
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathCAWrite,
		},

		HelpSynopsis: `Set the SSH private key used for signing certificates.`,
		HelpDescription: `This sets the CA information used for certificates generated by this
by this mount. The fields must be in the standard private and public SSH format.

For security reasons, the private key cannot be retrieved later.`,
	}
}

func (b *backend) pathCAWrite(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	publicKey := data.Get("public_key").(string)
	privateKey := data.Get("private_key").(string)

	_, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, errutil.UserError{Err: fmt.Sprintf(`Unable to parse "private_key" as an SSH private key: %s`, err)}
	}

	_, err = parsePublicSSHKey(publicKey)
	if err != nil {
		return nil, errutil.UserError{Err: fmt.Sprintf(`Unable to parse "public_key" as an SSH public key: %s`, err)}
	}

	err = req.Storage.Put(&logical.StorageEntry{
		Key:   "public_key",
		Value: []byte(publicKey),
	})
	if err != nil {
		return nil, err
	}

	bundle := signingBundle{
		Certificate: privateKey,
	}

	entry, err := logical.StorageEntryJSON("config/ca_bundle", bundle)
	if err != nil {
		return nil, err
	}

	err = req.Storage.Put(entry)
	return nil, err
}