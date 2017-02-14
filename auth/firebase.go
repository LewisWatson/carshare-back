package auth

import (
	"log"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
)

var (
	pem1 = "-----BEGIN CERTIFICATE-----\nMIIDHDCCAgSgAwIBAgIITHbjmfxnUfIwDQYJKoZIhvcNAQEFBQAwMTEvMC0GA1UE\nAxMmc2VjdXJldG9rZW4uc3lzdGVtLmdzZXJ2aWNlYWNjb3VudC5jb20wHhcNMTcw\nMjA5MDA0NTI2WhcNMTcwMjEyMDExNTI2WjAxMS8wLQYDVQQDEyZzZWN1cmV0b2tl\nbi5zeXN0ZW0uZ3NlcnZpY2VhY2NvdW50LmNvbTCCASIwDQYJKoZIhvcNAQEBBQAD\nggEPADCCAQoCggEBANZxLGwx2900shOdWTtOOevI/XmNbZWBBgC3hXWwcbuMHvW0\nqE0EAsY9EKxshQqQKOnRHWl0UzN+Qs5TcMhjaU12b0t8XHjz9GQls54dmMShQ3D9\nztjR57vIKcefrewrBtfyv0NuDnIvvAE4GZ0ikFEs5BFhd7EXL1t/D+quySQTKRTK\nzvU8l1fUDHLkSO11iWromzxN/WVNyMIHji7DDCXPshynmuJ7yxRHkFEmHOnshLdl\nLMzal5X798GP7zh0xL/xtIsJAddkiwLfw9BLtNZVU7t9EGCKhW0SQIRLRc6e6Jun\n9Y0M3nfmPaTirkRV//N0FRyQHuJWPcyEOcV7AQMCAwEAAaM4MDYwDAYDVR0TAQH/\nBAIwADAOBgNVHQ8BAf8EBAMCB4AwFgYDVR0lAQH/BAwwCgYIKwYBBQUHAwIwDQYJ\nKoZIhvcNAQEFBQADggEBAK190F6oI2lwmBEyWbLXEJbraLf/+nZoPReuYPxcU6/M\nGYg9GYvt7/cyKq4cTZMzzT6AeNuTOy153JYTGNNhkgsy+JyPH55jJxyJommjE+Xt\nt3k6KcJcOLeaVaOeJnFJfO1/pJAsVUTSJV2+U3u/Cjdn63vcsKmCmIye9iFdZCZ4\n997HPKyBgDFVBzUEHOO4qEWyyVEBEvmhEy0WVa6SX4XFvlgmIRa/bIdOA9ReFEwC\nlMD6pn0kEjAnNCm232Zl0cxW55CNx9wsaUOaBZSkppqYPZ+tgVmc5TS4bD/XsdY5\ntOCEJuTVqetQo8MLxW81yWlkAlSiJZJUKX2ShsL/efc=\n-----END CERTIFICATE-----\n"
	pem2 = "-----BEGIN CERTIFICATE-----\nMIIDHDCCAgSgAwIBAgIINc4DA5U/bREwDQYJKoZIhvcNAQEFBQAwMTEvMC0GA1UE\nAxMmc2VjdXJldG9rZW4uc3lzdGVtLmdzZXJ2aWNlYWNjb3VudC5jb20wHhcNMTcw\nMjEwMDA0NTI2WhcNMTcwMjEzMDExNTI2WjAxMS8wLQYDVQQDEyZzZWN1cmV0b2tl\nbi5zeXN0ZW0uZ3NlcnZpY2VhY2NvdW50LmNvbTCCASIwDQYJKoZIhvcNAQEBBQAD\nggEPADCCAQoCggEBALbS6/U33ln8h8TGv44RiO5684C2ach6n2a5GYYwERnEwzgS\n+gWACick+gXMTO+uNn3ROQfTHtQb99IkGj6DB9eTtiegdjwogNeRvxPC3hijwuaW\nFjxvqGByZfd6/f6CfAnboIR6bFlW7abWc2vZl0DwatCw4m5ldGPOJ0iIShqHYIad\nEdaQfjiElTMxcJ+4BHh0m+0DWPp3qc8HV3MwqqiP9TlW+bNGfRf8PW/RP2WIosBu\nrzHroZgVmWqnr2LGx9KKKcjiRcucdSfFRvDHY8qf8XkvxesbcvOsSVQQfXmU9Bhw\n7a38YDtOv20gQlZkHpwaG8QuJ2tzuetsr60stnUCAwEAAaM4MDYwDAYDVR0TAQH/\nBAIwADAOBgNVHQ8BAf8EBAMCB4AwFgYDVR0lAQH/BAwwCgYIKwYBBQUHAwIwDQYJ\nKoZIhvcNAQEFBQADggEBAHLXsbZCOC5oXlN41mUmeWQilbE9Ydf9n3Zle7P04RGn\nRKCxK8CUhHUU+qepKD2Yn9TUYyTELbGrgZcPu0G6p3Nrpk8ZnkiB33TYQPNSOL2o\nBWlV/YgShSt6rG746bSYHqCrW9OnThRE/jhfImEg1+nffqrQvtDAgQyouQtaOSp1\ncAHqWMPz03d9QkkFoDPoveVQLFo4zApdeaHqL0dIOAh1NPMvUeQzjujrgt5F5rgZ\nnGUoWjRyVVsWaylhXOL4sSLjtuk1NOTMo4C8S2kRHoGBIjvsUsG1ZZ+UFBKbWPMY\nJ92kwnisx6WUfbwK0YMeshO0/tbVLtLYOP+v9ZN/UdE=\n-----END CERTIFICATE-----\n"
	pem3 = "-----BEGIN CERTIFICATE-----\nMIIDHDCCAgSgAwIBAgIIZ36AHgMyvnQwDQYJKoZIhvcNAQEFBQAwMTEvMC0GA1UE\nAxMmc2VjdXJldG9rZW4uc3lzdGVtLmdzZXJ2aWNlYWNjb3VudC5jb20wHhcNMTcw\nMjA4MDA0NTI2WhcNMTcwMjExMDExNTI2WjAxMS8wLQYDVQQDEyZzZWN1cmV0b2tl\nbi5zeXN0ZW0uZ3NlcnZpY2VhY2NvdW50LmNvbTCCASIwDQYJKoZIhvcNAQEBBQAD\nggEPADCCAQoCggEBANBNTpiQplOYizNeLbs+r941T392wiuMWr1gSJEVykFyj7fe\nCCIhS/zrmG9jxVMK905KwceO/FNB4SK+l8GYLb559xZeJ6MFJ7QmRfL7Fjkq7GHS\n0/sOFpjX7vfKjxH5oT65Fb1+Hb4RzdoAjx0zRHkDIHIMiRzV0nYleplqLJXOAc6E\n5HQros8iLdf+ASdqaN0hS0nU5aa/cPu/EHQwfbEgYraZLyn5NtH8SPKIwZIeM7Fr\nnh+SS7JSadsqifrUBRtb//fueZ/FYlWqHEppsuIkbtaQmTjRycg35qpVSEACHkKc\nW05rRsSvz7q1Hucw6Kx/dNBBbkyHrR4Mc/wg31kCAwEAAaM4MDYwDAYDVR0TAQH/\nBAIwADAOBgNVHQ8BAf8EBAMCB4AwFgYDVR0lAQH/BAwwCgYIKwYBBQUHAwIwDQYJ\nKoZIhvcNAQEFBQADggEBAEuYEtvmZ4uReMQhE3P0iI4wkB36kWBe1mZZAwLA5A+U\niEODMVKaaCGqZXrJTRhvEa20KRFrfuGQO7U3FgOMyWmX3drl40cNZNb3Ry8rsuVi\nR1dxy6HpC39zba/DsgL07enZPMDksLRNv0dVZ/X/wMrTLrwwrglpCBYUlxGT9RrU\nf8nAwLr1E4EpXxOVDXAX8bNBl3TCb2fu6DT62ZSmlJV40K+wTRUlCqIewzJ0wMt6\nO8+6kVdgZH4iKLi8gVjdcFfNsEpbOBoZqjipJ63l4A3mfxOkma0d2XgKR12KAfYX\ncAVPgihAPoNoUPJK0Nj+CmvNlUBXCrl9TtqGjK7AKi8=\n-----END CERTIFICATE-----\n"
)

// Firebase module to verify and extract information from firebase JWT tokens
type Firebase struct{}

// Verify to satisfy the auth.Verify interface
func (fa Firebase) Verify(accessToken string) (jwt.Claims, error) {

	if accessToken == "" {
		return nil, ErrNilToken
	}

	token, err := jws.ParseJWT([]byte(accessToken))
	if err != nil {
		return nil, err
	}

	user := token.Claims().Get("user_id")

	log.Printf("user %v", user)

	rsaPublic1, _ := crypto.ParseRSAPublicKeyFromPEM([]byte(pem1))
	rsaPublic2, _ := crypto.ParseRSAPublicKeyFromPEM([]byte(pem2))
	rsaPublic3, _ := crypto.ParseRSAPublicKeyFromPEM([]byte(pem3))

	// Validate token
	err = token.Validate(rsaPublic1, crypto.SigningMethodRS256)
	if err == crypto.ErrECDSAVerification {
		err = token.Validate(rsaPublic2, crypto.SigningMethodRS256)
		if err == crypto.ErrECDSAVerification {
			err = token.Validate(rsaPublic3, crypto.SigningMethodRS256)
		}
	}

	switch err {
	case jwt.ErrTokenIsExpired:
		err = ErrTokenExpired
		break
	case crypto.ErrECDSAVerification:
		err = ErrECDSAVerification
		break
	case jws.ErrNotCompact:
		err = ErrNotCompact
		break
	}

	return token.Claims(), err
}
