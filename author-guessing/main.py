import random
import sqlite3
from collections import defaultdict
from typing import NamedTuple

from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity


class QuoteRow(NamedTuple):
    id: int
    created_at: str
    updated_at: str
    deleted_at: str | None
    text: str
    source: str | None
    elo: int
    author_id: str
    author_quote_id: int


db = sqlite3.connect("quotes.sqlite")

all_quotes = db.execute(
    """
    SELECT
	    *
    FROM
	    quotes
	    INNER JOIN quote_authors ON quote_authors.quote_id = quotes.id
    WHERE
	    quotes.deleted_at IS NULL
	    AND quotes.id IN (
		    SELECT
			    quote_id
		    FROM
			    quote_authors
		    GROUP BY
			    quote_id
		    HAVING
			    COUNT(*) = 1
	    );
    """
)

author_quotes = defaultdict(list)
quote_objs = [QuoteRow(*row) for row in all_quotes.fetchall()]

for quote in quote_objs:
    author_quotes[quote.author_id].append(quote)

author_quote_text = {
    author_id: "\n".join(q.text for q in quotes)
    for author_id, quotes in author_quotes.items()
}

vectorizer = TfidfVectorizer()

tfidf_matrix = vectorizer.fit_transform(author_quote_text.values())
index_to_author = list(author_quote_text.keys())
feature_names = vectorizer.get_feature_names_out()


def find_possible_quote_authors(quote: str, n: int = 5) -> list[str]:
    quote_vector = vectorizer.transform([quote])
    similarities = cosine_similarity(quote_vector, tfidf_matrix)

    similar_indices = similarities.argsort()[0][-n:][::-1]
    similar_authors = [index_to_author[i] for i in similar_indices]
    return similar_authors


random_quote = random.choice(quote_objs)
print(random_quote.text)
possible_authors = find_possible_quote_authors(random_quote.text)
print("Possible authors:", possible_authors)
print("Actual author:", random_quote.author_id)
