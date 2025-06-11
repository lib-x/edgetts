package communicate

import (
	"crypto/sha256"
	"fmt"
	"github.com/lib-x/edgetts/internal/businessConsts"
	"time"
)

func generateWssEndpoint() string {
	return businessConsts.EdgeWssEndpoint +
		"&Sec-MS-GEC=" + generateSecMsGecToken() +
		"&Sec-MS-GEC-Version=" + generateSecMsGecVersion() +
		"&ConnectionId=" + generateConnectID()
}

func generateSecMsGecToken() string {
	now := time.Now().UTC()
	ticks := (now.Unix() + 11644473600) * 10000000
	ticks = ticks - (ticks % 3_000_000_000)

	strToHash := fmt.Sprintf("%d%s", ticks, businessConsts.TrustedClientToken)
	hash := sha256.New()
	hash.Write([]byte(strToHash))
	hexDig := fmt.Sprintf("%X", hash.Sum(nil))
	return hexDig
}

// generateSecMsGecVersion  Sec-MS-GEC-Version token
func generateSecMsGecVersion() string {
	return fmt.Sprintf("1-%s", businessConsts.ChromiumFllVersion)
}
