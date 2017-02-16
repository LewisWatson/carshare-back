package auth_test

import (
	. "github.com/LewisWatson/carshare-back/auth"

	"github.com/SermoDigital/jose/jwt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Firebase", func() {

	var (
		firebase TokenVerifier
		// expiredToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjllYjY1NGE3YTNmYTJiMWQ5MzJmZGRhNWQ0YjVhY2NiODU5OGU4YmEifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vcmlkZXNoYXJlbG9nZ2VyIiwicHJvdmlkZXJfaWQiOiJhbm9ueW1vdXMiLCJhdWQiOiJyaWRlc2hhcmVsb2dnZXIiLCJhdXRoX3RpbWUiOjE0ODY3MTM3ODIsInVzZXJfaWQiOiI3MGRYckw1T3dKTWdJdmcxbUZ3STltVXVxdDAyIiwic3ViIjoiNzBkWHJMNU93Sk1nSXZnMW1Gd0k5bVV1cXQwMiIsImlhdCI6MTQ4NjcxMzc4MiwiZXhwIjoxNDg2NzE3MzgyLCJmaXJlYmFzZSI6eyJpZGVudGl0aWVzIjp7fSwic2lnbl9pbl9wcm92aWRlciI6ImFub255bW91cyJ9fQ.um1CgIWMRJbEFz61s8NOOEEmgO_qMP93Br1JqDiPtWR0JXcU4-nyGQL0xQkEAdVqAIJRA8asKkfdXzwQB4EWz604jX6JPPabIS8zOHvxgDf20KS_e7ZUSvlo2j0ZAwfRD0i6Uu-grQ1BxDksoY5lr_nxjjL9tQ86mtAHiNqa9gb4FfDlmqX-lq_Xzmyg6WOU6bUdj7Z4aDJgOoy4ZkX_TYuIjSJhdhqre8-7-Deb-pwu94L6h5qTuQpgmcZT4BFAcvMCTV3IFguc92jquoarfwIYtdw4aMnfhJHVNmUXCGt0jWCtBWLaUVpCzQDWfKSfDuBhKtG9zb69fyKCuXglvA"
		// validToken   = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjU3M2YxZGJlNTE4YmQyMjM3ZmNkNWJhNGVkYWYzNzEwY2QyNjA5MjYifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vcmlkZXNoYXJlbG9nZ2VyIiwicHJvdmlkZXJfaWQiOiJhbm9ueW1vdXMiLCJhdWQiOiJyaWRlc2hhcmVsb2dnZXIiLCJhdXRoX3RpbWUiOjE0ODcwOTM5NDAsInVzZXJfaWQiOiJmWUVKaU5CNzB0YUhuTTdRdGh1SU9YNzh2OVoyIiwic3ViIjoiZllFSmlOQjcwdGFIbk03UXRodUlPWDc4djlaMiIsImlhdCI6MTQ4NzA5Mzk0MCwiZXhwIjoxNDg3MDk3NTQwLCJmaXJlYmFzZSI6eyJpZGVudGl0aWVzIjp7fSwic2lnbl9pbl9wcm92aWRlciI6ImFub255bW91cyJ9fQ.b6fyhibigRyfxfBITcHrSVgYR_JSoXjNhrUqtUpopwbgvzzl3oK7oneMW3HJsNKiHUQhuiiMN30X42zQQ8qsME5fdkSPu5go0qAPgACWj1E0K10DryfAmSRwkbYtpOiY9bA6fGUsHikUBQsUy8VnoajII22udcDfdQI37-cMIOAsmK7laDS4FODl_87xUqmbPK4KeSnWty_tgRD-IyiKpyIJi67pR7Z1k3cfnHI2ZpWiE4Dxaw4Yl0_mA4NvnyRQ9Q1nZ4hre827VK4KkhQah7VWlCZdc8HtTCRzbb5HU46VdVfGmBFq5_ennWnIcGmGCFlFlS8mfzwbPrUwDpr9kQ"
	)

	BeforeEach(func() {
		var err error

		// creating a new firebase involves an HTTP request so only do it once
		if firebase == nil {
			firebase, err = NewFirebase("ridesharelogger")
			Expect(err).ToNot(HaveOccurred())
		}
	})

	Describe("validate", func() {

		var (
			err    error
			claims jwt.Claims
		)

		// Context("valid token", func() {

		// 	BeforeEach(func() {
		// 		claims, err = firebase.Verify(validToken)
		// 	})

		// 	It("should not throw error", func() {
		// 		Expect(err).ToNot(HaveOccurred())
		// 	})

		// 	It("should return claims containing the expected user", func() {
		// 		Expect(claims).ToNot(BeNil())
		// 		Expect(claims.Get("user_id")).To(Equal("70dXrL5OwJMgIvg1mFwI9mUuqt02"))
		// 	})

		// })

		// Context("expired valid token", func() {

		// 	BeforeEach(func() {
		// 		claims, err = firebase.Verify(expiredToken)
		// 	})

		// 	It("should throw token expired error", func() {
		// 		Expect(err).To(HaveOccurred())
		// 		Expect(err).To(Equal(ErrTokenExpired))
		// 	})

		// 	It("should return claims containing the expected user", func() {
		// 		Expect(claims).ToNot(BeNil())
		// 		Expect(claims.Get("user_id")).To(Equal("70dXrL5OwJMgIvg1mFwI9mUuqt02"))
		// 	})

		// })

		Context("invalid token", func() {

			BeforeEach(func() {
				claims, err = firebase.Verify("invalid token")
			})

			It("should throw ErrNotCompact error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(ErrNotCompact))
			})

			It("should return nil claims", func() {
				Expect(claims).To(BeNil())
			})
		})

	})

})
