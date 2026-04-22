package edgetts

import "testing"

func TestOptionToInternalDefaults(t *testing.T) {
	opt := (&option{}).toInternalOption()
	opt.CheckAndApplyDefaultOption()
	if opt.Voice == "" {
		t.Fatal("expected default voice")
	}
	if opt.Pitch == "" || opt.Rate == "" || opt.Volume == "" {
		t.Fatal("expected default pitch/rate/volume")
	}
}

func TestOptionAliases(t *testing.T) {
	opt := &option{}
	WithHTTPProxy("http://127.0.0.1:8080")(opt)
	WithSOCKS5Proxy("127.0.0.1:1080", "u", "p")(opt)
	WithInsecureSkipVerify()(opt)
	if opt.HTTPProxy == "" || opt.SOCKS5Proxy == "" || !opt.IgnoreSSLVerification {
		t.Fatalf("unexpected option state: %+v", opt)
	}
}
