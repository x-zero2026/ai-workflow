.PHONY: build deploy clean validate local test

# Build all Lambda functions
build:
	sam build

# Deploy to AWS (first time use: make deploy-guided)
deploy:
	sam deploy

# Deploy with guided prompts (first time)
deploy-guided:
	sam deploy --guided

# Clean build artifacts
clean:
	rm -rf .aws-sam

# Validate SAM template
validate:
	sam validate

# Start local API Gateway
local:
	sam local start-api --env-vars env.json

# Test a specific function locally
# Usage: make test-function FUNCTION=ListWorkflowsFunction
test-function:
	sam local invoke $(FUNCTION) --env-vars env.json

# Full rebuild and deploy
redeploy: clean build deploy

# Show logs for a function
# Usage: make logs FUNCTION=ListWorkflowsFunction
logs:
	sam logs -n $(FUNCTION) --stack-name ai-workflow-backend --tail

# Show all available commands
help:
	@echo "Available commands:"
	@echo "  make build          - Build all Lambda functions"
	@echo "  make deploy         - Deploy to AWS"
	@echo "  make deploy-guided  - Deploy with guided prompts (first time)"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make validate       - Validate SAM template"
	@echo "  make local          - Start local API Gateway"
	@echo "  make test-function  - Test a function locally (FUNCTION=name)"
	@echo "  make redeploy       - Clean, build, and deploy"
	@echo "  make logs           - Show function logs (FUNCTION=name)"
