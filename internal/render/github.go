package render

import "regexp"

// githubPrefix matches the start of a GitHub URL: the scheme and the host in
// any case, and the owner and the repository as groups 1 and 2.
const githubPrefix = `^(?i:https?://github\.com)/([A-Za-z0-9_-]+)/([A-Za-z0-9_.-]+)`

// githubPatterns are the GitHub URLs that githubText shortens, each with the
// function that builds the text from the groups of its match.
var githubPatterns = []struct {
	re   *regexp.Regexp
	text func(m []string) string
}{
	{
		regexp.MustCompile(githubPrefix + `/(?:issues|pull|discussions)/([0-9]+)$`),
		func(m []string) string { return m[1] + "/" + m[2] + "#" + m[3] },
	},
	{
		regexp.MustCompile(githubPrefix + `/(?:issues|pull|discussions)/([0-9]+)#(?:issuecomment-|discussion_r|discussioncomment-)[0-9]+$`),
		func(m []string) string { return m[1] + "/" + m[2] + "#" + m[3] + " (comment)" },
	},
	{
		regexp.MustCompile(githubPrefix + `/pull/([0-9]+)#pullrequestreview-[0-9]+$`),
		func(m []string) string { return m[1] + "/" + m[2] + "#" + m[3] + " (review)" },
	},
	{
		regexp.MustCompile(githubPrefix + `/(?:commit|pull/[0-9]+/commits)/([0-9a-fA-F]{7,40})$`),
		func(m []string) string { return m[1] + "/" + m[2] + "@" + m[3][:7] },
	},
}

// githubText returns the text GitHub shows for the URL of an issue, a pull
// request, a discussion or a commit, such as owner/repo#12 or
// owner/repo@0123456, and false for any other URL.
func githubText(url string) (string, bool) {
	for _, p := range githubPatterns {
		if m := p.re.FindStringSubmatch(url); m != nil {
			return p.text(m), true
		}
	}
	return "", false
}
