package manager

// import (
// 	"errors"
// 	"net/url"
// 	"os"

// 	"github.com/tychonis/cyanotype/internal/digest"
// 	"github.com/tychonis/cyanotype/model"
// )

// func TrackItem(item *model.Item) error {
// 	for _, ref := range item.Reference {
// 		err := processReference(ref)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func processReference(ref *model.Reference) error {
// 	parsed, err := url.Parse(ref.URI)
// 	if err != nil {
// 		return err
// 	}
// 	switch parsed.Scheme {
// 	case "file":
// 		filepath := parsed.Opaque
// 		info, err := os.Stat(filepath)
// 		if err != nil {
// 			return err
// 		}
// 		ref.Size = info.Size()
// 		ref.LastModified = info.ModTime()
// 		ref.Digest, err = digest.SHA256FromFile(filepath)
// 		return err
// 	default:
// 		return errors.New("scheme not supported")
// 	}
// }
