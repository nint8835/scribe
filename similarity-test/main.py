import sqlite3
import struct

import sqlite_vec
from sentence_transformers import SentenceTransformer


def serialize_f32(vector: list[float]) -> bytes:
    return struct.pack("%sf" % len(vector), *vector)


db = sqlite3.connect("quotes.sqlite")
db.enable_load_extension(True)
sqlite_vec.load(db)
db.enable_load_extension(False)

db.execute("DROP TABLE IF EXISTS vec_items")
db.execute("CREATE VIRTUAL TABLE vec_items USING vec0(embedding float[384])")

model = SentenceTransformer("all-MiniLM-L6-v2")

quotes = db.execute("SELECT id, text FROM quotes").fetchall()
for quote_id, text in quotes:
    print(quote_id, text)
    embedding = model.encode(text)
    db.execute(
        "INSERT INTO vec_items (rowid, embedding) VALUES (?, ?)", (quote_id, embedding)
    )
db.commit()

target_quote = input("Enter a quote: ")
target_embedding = model.encode(target_quote)

results = db.execute(
    "SELECT rowid, distance FROM vec_items WHERE embedding MATCH ? ORDER BY distance LIMIT 5",
    [serialize_f32(target_embedding)],
).fetchall()

for rowid, distance in results:
    quote = db.execute("SELECT text FROM quotes WHERE id = ?", (rowid,)).fetchone()
    print(f"{distance:.3f}: {quote[0]}")

db.close()
