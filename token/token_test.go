package token_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cpu-entitlement-plugin/token"
	"code.cloudfoundry.org/cpu-entitlement-plugin/token/tokenfakes"
)

var _ = Describe("Token", func() {
	var (
		fakeGetToken       *tokenfakes.FakeGetToken
		tokenGetter        *TokenGetter
		tenMinutesToken    string
		twentyMinutesToken string
	)

	BeforeEach(func() {
		fakeGetToken = new(tokenfakes.FakeGetToken)

		var err error
		tenMinutesToken, err = aTokenExpiringIn(10 * time.Minute)
		Expect(err).NotTo(HaveOccurred())
		twentyMinutesToken, err = aTokenExpiringIn(20 * time.Minute)
		Expect(err).NotTo(HaveOccurred())

		fakeGetToken.ReturnsOnCall(0, tenMinutesToken, nil)
		fakeGetToken.ReturnsOnCall(1, twentyMinutesToken, nil)
	})

	JustBeforeEach(func() {
		tokenGetter = NewTokenGetter(fakeGetToken.Spy)
	})

	It("returns a token", func() {
		token, err := tokenGetter.Token()
		Expect(err).NotTo(HaveOccurred())
		Expect(token).To(Equal(tenMinutesToken))
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

	Context("when the token metadata is not padded", func() {
		BeforeEach(func() {
			fakeGetToken.ReturnsOnCall(0, "foo.e30", nil) // "{}"
		})

		It("does not fail", func() {
			_, err := tokenGetter.Token()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when the token metadata is badly encoded", func() {
		BeforeEach(func() {
			fakeGetToken.ReturnsOnCall(0, "foo.***", nil)
		})

		It("returns the error", func() {
			_, err := tokenGetter.Token()
			Expect(err).To(MatchError(ContainSubstring("failed to decode token from base64")))
		})
	})

	Context("when the token metadata is not in the valid JSON format", func() {
		BeforeEach(func() {
			fakeGetToken.ReturnsOnCall(0, "foo.ew", nil) // "{"
		})

		It("returns the error", func() {
			_, err := tokenGetter.Token()
			Expect(err).To(MatchError(ContainSubstring("invalid token")))
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
			expiredToken, err := anExpiredToken()
			Expect(err).NotTo(HaveOccurred())
			fakeGetToken.ReturnsOnCall(0, expiredToken, nil)
		})

		It("returns a new token", func() {
			token1, err := tokenGetter.Token()
			Expect(err).NotTo(HaveOccurred())

			token2, err := tokenGetter.Token()
			Expect(err).NotTo(HaveOccurred())

			Expect(token1).NotTo(Equal(token2))
		})
	})
})

func aTokenExpiringIn(duration time.Duration) (string, error) {
	return aTokenExpiringOn(time.Now().Add(duration))
}

func anExpiredToken() (string, error) {
	return aTokenExpiringOn(time.Now())
}

func aTokenExpiringOn(t time.Time) (string, error) {
	tokenMetadata := map[string]interface{}{
		"exp": t.Unix(),
	}

	tokenMetadataJson, err := json.Marshal(tokenMetadata)
	if err != nil {
		return "", err
	}

	encodedTokenMetadata := base64.URLEncoding.EncodeToString(tokenMetadataJson)

	return fmt.Sprintf("foo.%s.bar", encodedTokenMetadata), nil
}
