# RAG (Vector Search) Setup Guide

## 概述

AI Workflow Center 支持使用向量搜索（RAG）来帮助 AI 助手找到最相关的工作流。这个功能是**可选的**，核心功能不依赖它。

---

## 前置条件

1. ✅ 已运行 `schema.sql` 或 `schema-simple.sql`
2. ✅ 已创建至少一些工作流数据
3. ✅ 有 OpenAI API Key（用于生成 embeddings）
4. ✅ pgvector 扩展已启用

---

## 步骤 1: 检查 pgvector 版本

在 Supabase SQL Editor 中运行：

```sql
SELECT extversion FROM pg_extension WHERE extname = 'vector';
```

**版本说明**:
- **0.5.0+**: 支持 HNSW 索引（推荐）
- **0.4.0+**: 支持 IVFFlat 索引
- **< 0.4.0**: 需要升级 pgvector

---

## 步骤 2: 生成 Embeddings

### 方法 A: 使用 Python 脚本（推荐）

创建 `generate_embeddings.py`:

```python
import os
import psycopg2
from openai import OpenAI

# 配置
OPENAI_API_KEY = "your-openai-api-key"
SUPABASE_URL = "https://rbpsksuuvtzmathnmyxn.supabase.co"
DB_PASSWORD = "iPass4xz2026!"

# 连接数据库
conn = psycopg2.connect(
    host="aws-0-ap-southeast-1.pooler.supabase.com",
    port=6543,
    database="postgres",
    user="postgres.rbpsksuuvtzmathnmyxn",
    password=DB_PASSWORD
)

# 初始化 OpenAI
client = OpenAI(api_key=OPENAI_API_KEY)

# 获取需要生成向量的工作流
cursor = conn.cursor()
cursor.execute("""
    SELECT workflow_id, workflow_name, description, parameters
    FROM workflows
    WHERE embedding IS NULL
    LIMIT 100
""")

workflows = cursor.fetchall()
print(f"Found {len(workflows)} workflows without embeddings")

# 批量生成 embeddings
for workflow_id, name, description, parameters in workflows:
    # 组合文本
    text = f"{name}\n{description}\n{parameters}"
    
    # 生成 embedding
    response = client.embeddings.create(
        model="text-embedding-3-small",  # 1536 dimensions
        input=text
    )
    
    embedding = response.data[0].embedding
    
    # 更新数据库
    cursor.execute("""
        UPDATE workflows
        SET embedding = %s,
            embedding_model = 'text-embedding-3-small',
            embedding_version = content_version::text
        WHERE workflow_id = %s
    """, (embedding, workflow_id))
    
    print(f"Generated embedding for: {name}")

conn.commit()
cursor.close()
conn.close()

print("Done! All embeddings generated.")
```

运行脚本:
```bash
pip install openai psycopg2-binary
python generate_embeddings.py
```

### 方法 B: 使用 SQL + Supabase Edge Function

如果你想在 Supabase 内部处理，可以创建 Edge Function。

---

## 步骤 3: 创建向量索引

在 Supabase SQL Editor 中运行 `create-vector-index.sql`:

```sql
-- 检查数据量
SELECT COUNT(*) FROM workflows WHERE embedding IS NOT NULL;

-- 如果有数据，创建索引
-- Option 1: HNSW (推荐，需要 pgvector 0.5.0+)
CREATE INDEX idx_workflows_embedding_hnsw ON workflows 
USING hnsw (embedding vector_cosine_ops) 
WITH (m = 16, ef_construction = 64);

-- Option 2: IVFFlat (兼容旧版本)
-- CREATE INDEX idx_workflows_embedding_ivfflat ON workflows 
-- USING ivfflat (embedding vector_cosine_ops) 
-- WITH (lists = 100);
```

**索引创建时间**:
- 100 条记录: < 1 秒
- 1,000 条记录: < 10 秒
- 10,000 条记录: < 1 分钟

---

## 步骤 4: 测试向量搜索

### 测试 SQL 函数

```sql
-- 生成查询的 embedding（需要先在应用中生成）
-- 假设查询是 "customer feedback workflow"
-- 对应的 embedding 是 [0.1, 0.2, 0.3, ...]

SELECT 
    workflow_id,
    workflow_name,
    description,
    1 - (embedding <=> '[0.1, 0.2, ...]'::vector(1536)) as similarity
FROM workflows
WHERE embedding IS NOT NULL
ORDER BY embedding <=> '[0.1, 0.2, ...]'::vector(1536)
LIMIT 5;
```

### 使用数据库函数

```sql
-- 使用预定义的搜索函数
SELECT * FROM search_workflows_vector(
    '[0.1, 0.2, ...]'::vector(1536),  -- query embedding
    'your-project-id'::uuid,           -- project_id
    5,                                  -- top_k
    0.7                                 -- threshold
);
```

---

## 步骤 5: 集成到应用

### 创建搜索 Lambda 函数

创建 `cmd/search-workflows/main.go`:

```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "log"
    "os"
    "net/http"
    "bytes"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/xzero/ai-workflow/pkg/auth"
    "github.com/xzero/ai-workflow/pkg/db"
    "github.com/xzero/ai-workflow/pkg/response"
)

var database *sql.DB

func init() {
    var err error
    database, err = db.Connect(
        os.Getenv("SUPABASE_URL"),
        os.Getenv("DB_PASSWORD"),
    )
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
}

type SearchRequest struct {
    Query     string `json:"query"`
    ProjectID string `json:"project_id"`
    TopK      int    `json:"top_k"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    // Validate token
    token, err := auth.ExtractToken(request.Headers["Authorization"])
    if err != nil {
        return response.Unauthorized("Invalid authorization header"), nil
    }

    claims, err := auth.ValidateToken(token, os.Getenv("JWT_SECRET"))
    if err != nil {
        return response.Unauthorized("Invalid or expired token"), nil
    }

    // Parse request
    var req SearchRequest
    if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
        return response.BadRequest("Invalid request body"), nil
    }

    // Generate embedding for query
    embedding, err := generateEmbedding(req.Query)
    if err != nil {
        log.Printf("Error generating embedding: %v", err)
        // Fallback to full-text search
        return searchFullText(req.Query, req.ProjectID, req.TopK)
    }

    // Vector search
    results, err := searchVector(embedding, req.ProjectID, req.TopK)
    if err != nil {
        log.Printf("Error in vector search: %v", err)
        // Fallback to full-text search
        return searchFullText(req.Query, req.ProjectID, req.TopK)
    }

    return response.Success(map[string]interface{}{
        "results":       results,
        "search_method": "vector",
    }), nil
}

func generateEmbedding(text string) ([]float64, error) {
    // Call OpenAI API
    apiKey := os.Getenv("OPENAI_API_KEY")
    
    reqBody := map[string]interface{}{
        "model": "text-embedding-3-small",
        "input": text,
    }
    
    jsonData, _ := json.Marshal(reqBody)
    
    req, _ := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Data []struct {
            Embedding []float64 `json:"embedding"`
        } `json:"data"`
    }
    
    json.NewDecoder(resp.Body).Decode(&result)
    
    return result.Data[0].Embedding, nil
}

func searchVector(embedding []float64, projectID string, topK int) ([]interface{}, error) {
    // Convert embedding to PostgreSQL vector format
    embeddingJSON, _ := json.Marshal(embedding)
    
    query := `
        SELECT * FROM search_workflows_vector(
            $1::vector(1536),
            $2::uuid,
            $3,
            0.7
        )
    `
    
    rows, err := database.Query(query, string(embeddingJSON), projectID, topK)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var results []interface{}
    // ... parse results
    
    return results, nil
}

func searchFullText(query, projectID string, topK int) (events.APIGatewayProxyResponse, error) {
    // Use full-text search as fallback
    // ...
    return response.Success(map[string]interface{}{
        "results":       []interface{}{},
        "search_method": "fulltext",
    }), nil
}

func main() {
    lambda.Start(handler)
}
```

---

## 性能优化

### 索引参数调优

**HNSW 参数**:
```sql
-- 默认（平衡）
CREATE INDEX ON workflows USING hnsw (embedding vector_cosine_ops) 
WITH (m = 16, ef_construction = 64);

-- 更快构建，稍低精度
CREATE INDEX ON workflows USING hnsw (embedding vector_cosine_ops) 
WITH (m = 8, ef_construction = 32);

-- 更高精度，较慢构建
CREATE INDEX ON workflows USING hnsw (embedding vector_cosine_ops) 
WITH (m = 32, ef_construction = 128);
```

**IVFFlat 参数**:
```sql
-- lists = sqrt(rows)
-- 1,000 rows: lists = 32
-- 10,000 rows: lists = 100
-- 100,000 rows: lists = 316

CREATE INDEX ON workflows USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);
```

---

## 维护

### 定期更新 Embeddings

创建定时任务（每小时运行）:

```python
# update_embeddings.py
# 查找 content_version > embedding_version 的记录
# 重新生成 embedding
```

### 监控

```sql
-- 检查有多少工作流有 embedding
SELECT 
    COUNT(*) as total,
    COUNT(embedding) as with_embedding,
    COUNT(*) - COUNT(embedding) as without_embedding
FROM workflows;

-- 检查过期的 embeddings
SELECT COUNT(*) 
FROM workflows 
WHERE content_version::text != embedding_version;
```

---

## 故障排查

### 问题 1: 索引创建失败

**错误**: `method hnsw does not exist`
- **解决**: 使用 IVFFlat 索引，或升级 pgvector

**错误**: `column cannot have more than 2000 dimensions`
- **解决**: 确保使用 1536 维（不是 2048）

### 问题 2: 搜索很慢

- 检查索引是否创建成功
- 增加 `lists` 参数（IVFFlat）
- 使用 HNSW 代替 IVFFlat

### 问题 3: 搜索结果不准确

- 检查 embedding 是否正确生成
- 调整相似度阈值
- 使用更好的 embedding 模型

---

## 成本估算

**OpenAI Embeddings API**:
- Model: text-embedding-3-small
- Cost: $0.02 / 1M tokens
- 估算: 每个工作流 ~100 tokens
- 1000 个工作流 ≈ $0.002

**存储**:
- 每个 embedding: 1536 * 4 bytes = 6KB
- 1000 个工作流 ≈ 6MB

---

## 总结

RAG 功能完全可选，但可以显著提升 AI 助手的体验：

1. ✅ 先部署核心功能（不需要 RAG）
2. ✅ 积累一些工作流数据
3. ✅ 按需启用 RAG 功能
4. ✅ 定期更新 embeddings

**推荐时机**: 当你有 100+ 工作流并且想要 AI 助手功能时再启用。
