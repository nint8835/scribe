package web

import (
	"fmt"
	"strings"

	"github.com/nint8835/scribe/pkg/database"
)

func (s *Server) formatAuthors(quote database.Quote) (string, error) {
	authorMentions := make([]string, len(quote.Authors))
	for i, author := range quote.Authors {
		authorMentions[i] = fmt.Sprintf("<@%s>", author.ID)
	}
	authorNames, err := s.resolveAuthorIDs(strings.Join(authorMentions, ", "))
	if err != nil {
		return "Invalid Authors", fmt.Errorf("error resolving author IDs: %w", err)
	}
	return authorNames, err
}
