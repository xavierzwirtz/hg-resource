package hg

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hg", func() {
	repo := Repository{
		Path: "/path/to/repo",
		Branch: "a_branch",
		IncludePaths: []string{
			"/path/1",
			"/path/2",
			"/path/3",
		},
		ExcludePaths: []string{
			"/path/4",
			"/path/5",
			"/path/6",
		},
	}
	emptyRepo := Repository{}
	Context("When making revset query fragments", func() {
		It("joins all given paths", func() {
			Expect(unionOfPaths(repo.IncludePaths)).To(Equal("file('re:/path/1')|file('re:/path/2')|file('re:/path/3')"))
		})

		It("produces the correct include fragment", func() {
			Expect(repo.makeIncludeQueryFragment()).To(Equal("file('re:/path/1')|file('re:/path/2')|file('re:/path/3')"))
			Expect(emptyRepo.makeIncludeQueryFragment()).To(Equal("all()"))
		})

		It("produces the correct exclude fragment", func() {
			Expect(repo.makeExcludeQueryFragment()).To(Equal("file('re:/path/4')|file('re:/path/5')|file('re:/path/6')"))
			Expect(emptyRepo.makeExcludeQueryFragment()).To(Equal("not all()"))
		})

	})
})
