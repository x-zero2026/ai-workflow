-- AI Workflow Center Database Schema (Simplified - No Vector Index)
-- Run this in your Supabase SQL editor

-- Enable pgvector extension for RAG support (optional)
CREATE EXTENSION IF NOT EXISTS vector;

-- Workflows table
CREATE TABLE IF NOT EXISTS workflows (
    workflow_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    source VARCHAR(50) NOT NULL CHECK (source IN ('coze', 'n8n')),
    template_name VARCHAR(50) NOT NULL CHECK (template_name IN ('workflow', 'streamflow')),
    
    -- HTTP request configuration
    http_method VARCHAR(10) NOT NULL CHECK (http_method IN ('GET', 'POST', 'PUT')),
    base_url VARCHAR(500) NOT NULL,
    bearer_token TEXT NOT NULL,
    external_workflow_id VARCHAR(255) NOT NULL,
    
    -- Parameters and headers (JSON)
    parameters JSONB DEFAULT '{}',
    headers JSONB DEFAULT '{}',
    
    -- Project and creator
    project_id UUID NOT NULL,
    creator_did VARCHAR(66) NOT NULL,
    
    -- Sharing status
    is_shared BOOLEAN DEFAULT FALSE,
    
    -- RAG support (for future AI assistant) - 1536 dimensions for OpenAI embeddings
    embedding vector(1536),
    embedding_model VARCHAR(100),
    embedding_version VARCHAR(50),
    content_version INTEGER DEFAULT 1,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for workflows
CREATE INDEX IF NOT EXISTS idx_workflows_project ON workflows(project_id);
CREATE INDEX IF NOT EXISTS idx_workflows_creator ON workflows(creator_did);
CREATE INDEX IF NOT EXISTS idx_workflows_shared ON workflows(is_shared);
CREATE INDEX IF NOT EXISTS idx_workflows_created ON workflows(created_at DESC);

-- Note: Vector index will be created later when needed for RAG functionality
-- You can create it manually when you have data:
-- CREATE INDEX idx_workflows_embedding ON workflows USING hnsw (embedding vector_cosine_ops);

-- Project workflow settings (for hiding workflows)
CREATE TABLE IF NOT EXISTS project_workflow_settings (
    project_id UUID NOT NULL,
    workflow_id UUID NOT NULL REFERENCES workflows(workflow_id) ON DELETE CASCADE,
    is_hidden BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, workflow_id)
);

CREATE INDEX IF NOT EXISTS idx_project_workflow_settings_project ON project_workflow_settings(project_id);

-- Trigger: Update updated_at on workflows
CREATE OR REPLACE FUNCTION update_workflows_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_workflows_updated_at
BEFORE UPDATE ON workflows
FOR EACH ROW
EXECUTE FUNCTION update_workflows_updated_at();

-- Trigger: Increment content_version when workflow content changes
CREATE OR REPLACE FUNCTION increment_workflow_content_version()
RETURNS TRIGGER AS $$
BEGIN
    IF (OLD.workflow_name != NEW.workflow_name OR 
        OLD.description != NEW.description OR 
        OLD.parameters != NEW.parameters OR
        OLD.headers != NEW.headers) THEN
        NEW.content_version = OLD.content_version + 1;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_increment_workflow_content_version
BEFORE UPDATE ON workflows
FOR EACH ROW
EXECUTE FUNCTION increment_workflow_content_version();

-- Trigger: Update updated_at on project_workflow_settings
CREATE OR REPLACE FUNCTION update_project_workflow_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_project_workflow_settings_updated_at
BEFORE UPDATE ON project_workflow_settings
FOR EACH ROW
EXECUTE FUNCTION update_project_workflow_settings_updated_at();

-- Function: Get visible workflows for a project
CREATE OR REPLACE FUNCTION get_visible_workflows(p_project_id UUID, p_user_did VARCHAR(66))
RETURNS TABLE (
    workflow_id UUID,
    workflow_name VARCHAR(255),
    description TEXT,
    source VARCHAR(50),
    template_name VARCHAR(50),
    http_method VARCHAR(10),
    base_url VARCHAR(500),
    bearer_token TEXT,
    external_workflow_id VARCHAR(255),
    parameters JSONB,
    headers JSONB,
    project_id UUID,
    creator_did VARCHAR(66),
    is_shared BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        w.workflow_id,
        w.workflow_name,
        w.description,
        w.source,
        w.template_name,
        w.http_method,
        w.base_url,
        w.bearer_token,
        w.external_workflow_id,
        w.parameters,
        w.headers,
        w.project_id,
        w.creator_did,
        w.is_shared,
        w.created_at,
        w.updated_at
    FROM workflows w
    LEFT JOIN project_workflow_settings pws 
        ON w.workflow_id = pws.workflow_id 
        AND pws.project_id = p_project_id
    WHERE (
        -- Project's own workflows
        w.project_id = p_project_id
        OR
        -- Shared workflows from other projects
        w.is_shared = true
    )
    AND (
        -- Not hidden
        pws.is_hidden IS NULL OR pws.is_hidden = false
    )
    ORDER BY w.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- Function: Search workflows using vector similarity (for RAG)
-- Note: This requires vector index to be created first
CREATE OR REPLACE FUNCTION search_workflows_vector(
    p_query_embedding vector(1536),
    p_project_id UUID,
    p_top_k INTEGER DEFAULT 5,
    p_threshold FLOAT DEFAULT 0.7
)
RETURNS TABLE (
    workflow_id UUID,
    workflow_name VARCHAR(255),
    description TEXT,
    source VARCHAR(50),
    template_name VARCHAR(50),
    similarity FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        w.workflow_id,
        w.workflow_name,
        w.description,
        w.source,
        w.template_name,
        1 - (w.embedding <=> p_query_embedding) AS similarity
    FROM workflows w
    LEFT JOIN project_workflow_settings pws 
        ON w.workflow_id = pws.workflow_id 
        AND pws.project_id = p_project_id
    WHERE (
        w.project_id = p_project_id OR w.is_shared = true
    )
    AND (pws.is_hidden IS NULL OR pws.is_hidden = false)
    AND w.embedding IS NOT NULL
    AND (1 - (w.embedding <=> p_query_embedding)) >= p_threshold
    ORDER BY similarity DESC
    LIMIT p_top_k;
END;
$$ LANGUAGE plpgsql;

-- Function: Search workflows using full-text search (fallback for RAG)
CREATE OR REPLACE FUNCTION search_workflows_fulltext(
    p_query TEXT,
    p_project_id UUID,
    p_top_k INTEGER DEFAULT 5
)
RETURNS TABLE (
    workflow_id UUID,
    workflow_name VARCHAR(255),
    description TEXT,
    source VARCHAR(50),
    template_name VARCHAR(50),
    relevance FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        w.workflow_id,
        w.workflow_name,
        w.description,
        w.source,
        w.template_name,
        ts_rank(
            to_tsvector('english', w.workflow_name || ' ' || w.description),
            plainto_tsquery('english', p_query)
        ) AS relevance
    FROM workflows w
    LEFT JOIN project_workflow_settings pws 
        ON w.workflow_id = pws.workflow_id 
        AND pws.project_id = p_project_id
    WHERE (
        w.project_id = p_project_id OR w.is_shared = true
    )
    AND (pws.is_hidden IS NULL OR pws.is_hidden = false)
    AND (
        to_tsvector('english', w.workflow_name || ' ' || w.description) @@ 
        plainto_tsquery('english', p_query)
    )
    ORDER BY relevance DESC
    LIMIT p_top_k;
END;
$$ LANGUAGE plpgsql;

-- Comments
COMMENT ON TABLE workflows IS 'AI工作流配置表';
COMMENT ON TABLE project_workflow_settings IS '项目工作流设置表（用于隐藏工作流）';
COMMENT ON COLUMN workflows.embedding IS '工作流向量表示（用于RAG搜索，1536维适配OpenAI embeddings）';
COMMENT ON COLUMN workflows.content_version IS '内容版本号，内容变更时自动递增';
COMMENT ON COLUMN workflows.embedding_version IS '向量版本号，用于追踪向量是否需要更新';

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'AI Workflow Center database schema created successfully!';
    RAISE NOTICE 'Tables: workflows, project_workflow_settings';
    RAISE NOTICE 'Note: Vector index is commented out. Create it later when needed for RAG.';
END $$;
