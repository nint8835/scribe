import sqlite3
import struct

import sqlite_vec
from sentence_transformers import SentenceTransformer


def serialize_f32(vector: list[float]) -> bytes:
    return struct.pack("%sf" % len(vector), *vector)


model = SentenceTransformer("all-MiniLM-L6-v2")

db = sqlite3.connect("quotes.sqlite")
db.enable_load_extension(True)
sqlite_vec.load(db)
db.enable_load_extension(False)

db.execute("DROP TABLE IF EXISTS quote_embeddings")
db.execute(
    "CREATE VIRTUAL TABLE quote_embeddings USING vec0(embedding float[384] distance_metric=cosine)"
)

quotes = db.execute("SELECT id, text FROM quotes").fetchall()
for quote_id, text in quotes:
    print(quote_id, text)
    embedding = model.encode(text)
    db.execute(
        "INSERT INTO quote_embeddings (rowid, embedding) VALUES (?, ?)",
        (quote_id, embedding),
    )
db.commit()

target_quote = input("Enter a quote: ")
target_embedding = model.encode(target_quote)

results = db.execute(
    "SELECT rowid, distance, quotes.text FROM quote_embeddings LEFT JOIN quotes ON quotes.id = quote_embeddings.rowid WHERE quote_embeddings.embedding MATCH ? AND quote_embeddings.k=5 ORDER BY distance",
    [serialize_f32(target_embedding)],
).fetchall()
for rowid, distance, text in results:
    print(f"{rowid}: {distance:.4f} - {text}")

db.close()
