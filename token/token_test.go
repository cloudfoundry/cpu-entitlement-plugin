package token_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cpu-entitlement-plugin/token"
	"code.cloudfoundry.org/cpu-entitlement-plugin/token/tokenfakes"
)

var _ = Describe("Token", func() {
	var (
		tokenLifetime time.Duration
		fakeGetToken  *tokenfakes.FakeGetToken
		tokenGetter   *TokenGetter
	)

	BeforeEach(func() {
		tokenLifetime = time.Minute
		fakeGetToken = new(tokenfakes.FakeGetToken)
		fakeGetToken.ReturnsOnCall(0, "token", nil)
		fakeGetToken.ReturnsOnCall(1, "token-new", nil)
	})

	JustBeforeEach(func() {
		tokenGetter = NewTokenGetter(fakeGetToken.Spy, tokenLifetime)
	})

	It("returns a token", func() {
		token, err := tokenGetter.Token()
		Expect(err).NotTo(HaveOccurred())
		Expect(token).To(Equal("token"))
	})

	Context("when getting the token fails", func() {
		BeforeEach(func() {
			fakeGetToken.ReturnsOnCall(0, "", errors.New("get-token-failure"))
		})

		It("returns the error", func() {
			_, err := tokenGetter.Token()
			Expect(err).To(MatchError("get-token-failure"))
		})
	})

	Context("when the token lifetime has not expired", func() {
		It("returns the same token", func() {
			token1, err := tokenGetter.Token()
			Expect(err).NotTo(HaveOccurred())
			token2, err := tokenGetter.Token()
			Expect(err).NotTo(HaveOccurred())
			Expect(token1).To(Equal(token2))
		})
	})

	Context("when the token lifetime expires", func() {
		BeforeEach(func() {
			tokenLifetime = time.Millisecond
		})

		It("returns a new token", func() {
			token1, err := tokenGetter.Token()
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(2 * time.Millisecond)
			token2, err := tokenGetter.Token()
			Expect(err).NotTo(HaveOccurred())

			Expect(token1).NotTo(Equal(token2))
		})
	})
})
