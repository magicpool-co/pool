package hostpool

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/magicpool-co/pool/pkg/sshtunnel"
)

func parseURL(url string, port int, tunnel *sshtunnel.SSHTunnel) (string, string, error) {
	id := url
	url = fmt.Sprintf("%s:%d", url, port)

	// tunnel the host if required (and active tunnel exists)
	if len(url) < 9 {
		return "", "", fmt.Errorf("url too short: %s", url)
	} else if url[:9] == "tunnel://" {
		if tunnel == nil {
			return "", "", fmt.Errorf("no active tunnel to use")
		}

		var err error
		url, err = tunnel.AddDestination(url[9:])
		if err != nil {
			return "", "", err
		}
	}

	return url, id, nil
}

func generateID(input string) string {
	h := sha1.New()
	h.Write([]byte(input))

	return hex.EncodeToString(h.Sum(nil))
}
