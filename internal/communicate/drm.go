package communicate

import (
	"crypto/sha256"
	"fmt"
	"github.com/lib-x/edgetts/internal/businessConsts"
	"time"
)

func GenerateSecMsGecToken() string {
	now := time.Now().UTC()
	ticks := (now.Unix() + 11644473600) * 10000000
	ticks = ticks - (ticks % 3_000_000_000)

	strToHash := fmt.Sprintf("%d%s", ticks, businessConsts.TrustedClientToken)
	hash := sha256.New()
	hash.Write([]byte(strToHash))
	hexDig := fmt.Sprintf("%X", hash.Sum(nil))
	return hexDig
}

// GenerateSecMsGecVersion  Sec-MS-GEC-Version token
func GenerateSecMsGecVersion() string {
	return fmt.Sprintf("1-%s", businessConsts.ChromiumFllVersion)
}
