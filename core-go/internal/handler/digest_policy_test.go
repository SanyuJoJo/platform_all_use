package handler

import "testing"

func TestValidateDigestForAlgorithm(t *testing.T) {
	cases := []struct {
		alg, dg string
		ok      bool
	}{
		{"SM2", "SHA1", false},
		{"SM2", "SHA256", false},
		{"SM2", "SM3", true},
		{"RSA", "SHA1", true},
		{"RSA", "SHA224", true},
		{"RSA", "SHA256", true},
		{"RSA", "SHA384", true},
		{"RSA", "SHA512", true},
		{"RSA", "SM3", false},
		{"RSA2048", "SHA512", true},
		{"ECC", "SHA224", true},
		{"ECC", "SM3", false},
		{"SLH-DSA", "SHA256", false},
		{"ML-DSA", "SM3", false},
		{"Falcon", "SHA256", false},
		{"SM2", "", true},
	}
	for _, c := range cases {
		_, err := ValidateDigestForAlgorithm(c.alg, c.dg)
		if c.ok && err != nil {
			t.Errorf("alg=%s digest=%s expect ok, got %v", c.alg, c.dg, err)
		}
		if !c.ok && err == nil {
			t.Errorf("alg=%s digest=%s expect err, got nil", c.alg, c.dg)
		}
	}
}
