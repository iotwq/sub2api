package service

const (
	BasispointsIgnoreImagesKey           = "openai_basispoints_ignore_images"
	BasispointsIgnoreEncryptedContentKey = "openai_basispoints_ignore_encrypted_content"
)

func (a *Account) IsBasispointsIgnoreImagesEnabled() bool {
	return a.UsesBasispointsResponses() && a.Extra[BasispointsIgnoreImagesKey] == true
}

func (a *Account) IsBasispointsIgnoreEncryptedContentEnabled() bool {
	return a.UsesBasispointsResponses() && a.Extra[BasispointsIgnoreEncryptedContentKey] == true
}
