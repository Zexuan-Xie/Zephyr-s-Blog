# Treat hybrid retrieval as a core feature

Search is a showcase feature for the blog, so the first release will implement hybrid retrieval with PostgreSQL full-text search plus Qwen/DashScope `text-embedding-v4` semantic retrieval through pgvector, fused with Reciprocal Rank Fusion (RRF), instead of stopping at `ILIKE`. LLM-based query expansion and reranking, including DeepSeek or Qwen `qwen3-rerank`, are deliberately left behind future provider interfaces because the database-backed retrieval path is easier to ship and review under the deadline.
