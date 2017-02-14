package auth_test

import (
	. "github.com/LewisWatson/carshare-back/auth"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Firebase", func() {

	var (
		firebase   TokenVerifier
		validToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjllYjY1NGE3YTNmYTJiMWQ5MzJmZGRhNWQ0YjVhY2NiODU5OGU4YmEifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vcmlkZXNoYXJlbG9nZ2VyIiwicHJvdmlkZXJfaWQiOiJhbm9ueW1vdXMiLCJhdWQiOiJyaWRlc2hhcmVsb2dnZXIiLCJhdXRoX3RpbWUiOjE0ODY3MTM3ODIsInVzZXJfaWQiOiI3MGRYckw1T3dKTWdJdmcxbUZ3STltVXVxdDAyIiwic3ViIjoiNzBkWHJMNU93Sk1nSXZnMW1Gd0k5bVV1cXQwMiIsImlhdCI6MTQ4NjcxMzc4MiwiZXhwIjoxNDg2NzE3MzgyLCJmaXJlYmFzZSI6eyJpZGVudGl0aWVzIjp7fSwic2lnbl9pbl9wcm92aWRlciI6ImFub255bW91cyJ9fQ.um1CgIWMRJbEFz61s8NOOEEmgO_qMP93Br1JqDiPtWR0JXcU4-nyGQL0xQkEAdVqAIJRA8asKkfdXzwQB4EWz604jX6JPPabIS8zOHvxgDf20KS_e7ZUSvlo2j0ZAwfRD0i6Uu-grQ1BxDksoY5lr_nxjjL9tQ86mtAHiNqa9gb4FfDlmqX-lq_Xzmyg6WOU6bUdj7Z4aDJgOoy4ZkX_TYuIjSJhdhqre8-7-Deb-pwu94L6h5qTuQpgmcZT4BFAcvMCTV3IFguc92jquoarfwIYtdw4aMnfhJHVNmUXCGt0jWCtBWLaUVpCzQDWfKSfDuBhKtG9zb69fyKCuXglvA"
		// validToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJuYmYiOjE0NDQ0Nzg0MDB9.u1riaD1rW97opCoAuRCTy4w58Br-Zk-bh7vLiRIsrpU"
	)

	BeforeEach(func() {
		firebase = &Firebase{}
	})

	Describe("validate", func() {

		var (
			err error
		)

		Context("expired valid token", func() {

			BeforeEach(func() {
				err = firebase.Verify(validToken)
			})

			It("should throw token expired error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(ErrTokenExpired))
			})

		})

		Context("invalid token", func() {

			BeforeEach(func() {
				err = firebase.Verify("invalid token")
			})

			It("should throw ErrNotCompact error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(ErrNotCompact))
			})
		})

	})

})
