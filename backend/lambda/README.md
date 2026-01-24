# AI Workflow Center - Lambda Deployment

✅ **Build Status**: All Lambda functions built successfully!

This directory contains the AWS SAM configuration for deploying the AI Workflow Center backend to AWS Lambda.

---

## 🚀 Quick Start

### First Time Deployment

```bash
# 1. Build all Lambda functions
make build

# 2. Deploy with guided prompts
make deploy-guided
```

### Subsequent Deployments

```bash
# Build and deploy
make build
make deploy

# Or in one command
make redeploy
```

---

## 📁 Directory Structure

```
lambda/
├── template.yaml          # SAM template (all 7 Lambda functions)
├── env.json              # Environment variables (DO NOT COMMIT)
├── env.json.example      # Environment variables template
├── samconfig.toml        # SAM deployment configuration
├── samconfig.toml.example # SAM configuration template
├── Makefile              # Build and deployment commands
└── .aws-sam/             # Build artifacts (auto-generated)
```

---

## 🔧 Configuration

### Environment Variables (`env.json`)

Already configured with:
- Supabase URL: `https://rbpsksuuvtzmathnmyxn.supabase.co`
- Database password: `iPass4xz2026!`
- JWT secret: Shared with DID Login platform

### SAM Configuration (`samconfig.toml`)

Already configured with:
- Stack name: `ai-workflow-backend`
- Region: `us-east-1`
- Parameters: Supabase URL, DB password, JWT secret

---

## 📦 Lambda Functions

All functions use:
- Runtime: `provided.al2` (custom Go runtime)
- Architecture: `arm64` (better performance, lower cost)
- Memory: 128 MB
- Timeout: 30 seconds (execute-workflow: 60 seconds)

### Functions:

1. **ListWorkflowsFunction** - `GET /api/projects/{id}/workflows`
2. **CreateWorkflowFunction** - `POST /api/workflows`
3. **ExecuteWorkflowFunction** - `POST /api/workflows/{id}/execute`
4. **UpdateWorkflowFunction** - `PUT /api/workflows/{id}`
5. **DeleteWorkflowFunction** - `DELETE /api/workflows/{id}`
6. **ShareWorkflowFunction** - `PUT /api/workflows/{id}/share`
7. **HideWorkflowFunction** - `PUT /api/projects/{projectId}/workflows/{workflowId}/hide`

---

## 🛠️ Available Commands

```bash
make build          # Build all Lambda functions
make deploy         # Deploy to AWS
make deploy-guided  # Deploy with guided prompts (first time)
make clean          # Clean build artifacts
make validate       # Validate SAM template
make local          # Start local API Gateway
make test-function  # Test a function locally (FUNCTION=name)
make redeploy       # Clean, build, and deploy
make logs           # Show function logs (FUNCTION=name)
```

---

## 🧪 Testing

### Local API Testing

```bash
# Start local API Gateway
make local

# In another terminal, test endpoints
curl http://localhost:3000/api/projects/123/workflows \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 📊 Deployment Process

### What Happens During Deployment

1. **Build Phase**: Compiles Go code for Linux ARM64
2. **Package Phase**: Creates ZIP files and uploads to S3
3. **Deploy Phase**: Creates/updates CloudFormation stack

### Expected Output

```
Successfully created/updated stack - ai-workflow-backend in us-east-1

Outputs:
ApiGatewayUrl = https://abc123.execute-api.us-east-1.amazonaws.com/prod
```

---

## 🐛 Troubleshooting

### Build Fails

**Problem**: `go.mod not found`
```bash
cd ../go && go mod tidy
```

**Problem**: `missing go.sum entry`
```bash
cd ../go && go mod tidy
```

### Deployment Fails

**Problem**: `No S3 bucket specified`
```bash
make deploy-guided
```

### Function Errors

**Problem**: Database connection timeout
- Check Supabase URL is correct
- Verify using Pooler port 6543

**Problem**: JWT validation failed
- Verify JWT_SECRET matches DID Login
- Check token hasn't expired

---

## 📈 Monitoring

```bash
# View logs for specific function
make logs FUNCTION=ListWorkflowsFunction

# Or use AWS CLI
aws logs tail /aws/lambda/ai-workflow-backend-ListWorkflowsFunction --follow
```

---

## ✅ Deployment Checklist

Before deploying:
- [x] Database schema deployed to Supabase
- [x] `env.json` configured with correct values
- [x] `samconfig.toml` configured
- [ ] AWS CLI configured with credentials
- [ ] SAM CLI installed

After deploying:
- [ ] Save API Gateway URL
- [ ] Test all endpoints
- [ ] Update frontend `.env` with API URL

---

**Status**: ✅ Ready to deploy!

Run `make deploy-guided` to get started.
