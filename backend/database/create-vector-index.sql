-- Create Vector Index for RAG Search
-- Run this AFTER you have inserted workflow data (recommended: at least 100 rows)

-- Check pgvector version
SELECT extversion FROM pg_extension WHERE extname = 'vector';

-- Check if you have workflow data
SELECT COUNT(*) as workflow_count FROM workflows;

-- Option 1: HNSW Index (Recommended for pgvector 0.5.0+)
-- Better performance, supports higher dimensions
-- Requires: pgvector >= 0.5.0
CREATE INDEX IF NOT EXISTS idx_workflows_embedding_hnsw ON workflows 
USING hnsw (embedding vector_cosine_ops) 
WITH (m = 16, ef_construction = 64);

-- If Option 1 fails with "method hnsw does not exist", use Option 2 instead:

-- Option 2: IVFFlat Index (Compatible with older pgvector versions)
-- Works with pgvector 0.4.0+
-- Good for up to 1 million vectors
-- CREATE INDEX IF NOT EXISTS idx_workflows_embedding_ivfflat ON workflows 
-- USING ivfflat (embedding vector_cosine_ops) 
-- WITH (lists = 100);

-- Verify index was created
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'workflows' AND indexname LIKE '%embedding%';

-- Test vector search (requires embeddings to be populated first)
-- SELECT workflow_id, workflow_name, 
--        1 - (embedding <=> '[0.1, 0.2, ...]'::vector) as similarity
-- FROM workflows
-- WHERE embedding IS NOT NULL
-- ORDER BY embedding <=> '[0.1, 0.2, ...]'::vector
-- LIMIT 5;

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'Vector index creation completed!';
    RAISE NOTICE 'You can now use vector similarity search for RAG.';
END $$;
