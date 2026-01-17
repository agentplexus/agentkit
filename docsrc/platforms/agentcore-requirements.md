# AWS AgentCore Infrastructure Requirements

This document outlines the requirements and constraints for deploying agents to AWS Bedrock AgentCore. These apply to all IaC tools (CDK, Pulumi, Terraform, CloudFormation).

## Container Image Requirements

### ECR Only

AgentCore **only supports Amazon ECR** container images. Third-party registries like GHCR, Docker Hub, or GCR are not supported.

**Required format:**

```
{account_id}.dkr.ecr.{region}.amazonaws.com/{repository}:{tag}
```

**Examples:**

```
# Valid
123456789012.dkr.ecr.us-west-2.amazonaws.com/stats-agent-research:latest
123456789012.dkr.ecr.us-west-2.amazonaws.com/my-org/my-agent:v1.0.0

# Invalid - will fail validation
ghcr.io/myorg/my-agent:latest
docker.io/myimage:latest
gcr.io/my-project/my-agent:latest
```

### Migrating from GHCR to ECR

If your images are in GHCR, you need to copy them to ECR:

```bash
# Create ECR repository
aws ecr create-repository --repository-name stats-agent-research --region us-west-2

# Login to both registries
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 123456789012.dkr.ecr.us-west-2.amazonaws.com
echo $GHCR_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Pull from GHCR
docker pull ghcr.io/myorg/stats-agent-research:latest

# Tag for ECR
docker tag ghcr.io/myorg/stats-agent-research:latest 123456789012.dkr.ecr.us-west-2.amazonaws.com/stats-agent-research:latest

# Push to ECR
docker push 123456789012.dkr.ecr.us-west-2.amazonaws.com/stats-agent-research:latest
```

## Runtime Configuration

### Endpoint Naming

Endpoint names must match the pattern `^[a-zA-Z][a-zA-Z0-9_]{0,47}$`:

- Must start with a letter
- Can contain letters, numbers, and underscores only
- **No hyphens allowed**
- Maximum 48 characters

**Examples:**

```
# Valid
research_endpoint
synthesisEndpoint
agent1_endpoint

# Invalid
research-endpoint    # hyphens not allowed
1_endpoint           # must start with letter
my-agent-endpoint    # hyphens not allowed
```

### Timeout (MaxLifetime)

The `MaxLifetime` (timeout) must be **at least 60 seconds**.

```json
{
  "timeoutSeconds": 60   // Minimum value
}
```

**Note:** AgentCore supports sessions up to 8 hours (28800 seconds).

### Memory Allocation

Valid memory values in MB:

- 512
- 1024
- 2048
- 4096
- 8192
- 16384

## Gateway Configuration

### Protocol Type

The Gateway **only supports MCP protocol**. HTTP is not a valid option for Gateway.

```json
{
  "gateway": {
    "enabled": true,
    "protocol": "MCP"  // Only valid option
  }
}
```

**Note:** Individual agent runtimes can use HTTP protocol, but the Gateway resource itself only supports MCP.

### Authorizer Type

Valid authorizer types:

- `NONE` - No authorization (default)
- `IAM` - AWS IAM authorization
- `CUSTOM_JWT` - Custom JWT authorizer

## Network Configuration

### VPC Requirements

AgentCore runtimes require VPC configuration with:

- Private subnets (for agent execution)
- Security groups allowing inter-agent communication
- VPC endpoints recommended for:
  - ECR (ecr.api, ecr.dkr)
  - Secrets Manager
  - CloudWatch Logs
  - S3 (gateway endpoint for ECR layers)
  - Bedrock (if using Bedrock models)

### Network Mode

Currently only `VPC` network mode is supported.

## IAM Requirements

The execution role must have permissions for:

- ECR image pull (`ecr:GetAuthorizationToken`, `ecr:BatchGetImage`, etc.)
- CloudWatch Logs (`logs:CreateLogStream`, `logs:PutLogEvents`)
- Secrets Manager (if using secrets)
- Bedrock (if invoking Bedrock models)

## Regional Availability

AgentCore Runtime is available in these regions (as of January 2025):

- US East (N. Virginia) - us-east-1
- US East (Ohio) - us-east-2
- US West (Oregon) - us-west-2
- Europe (Frankfurt) - eu-central-1
- Europe (Ireland) - eu-west-1
- Asia Pacific (Mumbai) - ap-south-1
- Asia Pacific (Singapore) - ap-southeast-1
- Asia Pacific (Sydney) - ap-southeast-2
- Asia Pacific (Tokyo) - ap-northeast-1

## Configuration Checklist

Before deploying, verify:

- [ ] Container images are in ECR (not GHCR/Docker Hub)
- [ ] Endpoint names use underscores, not hyphens
- [ ] Timeout is at least 60 seconds
- [ ] Memory is a valid value (512, 1024, 2048, 4096, 8192, 16384)
- [ ] Gateway protocol is MCP (if using Gateway)
- [ ] VPC has required subnets and endpoints
- [ ] IAM role has necessary permissions
- [ ] Deploying to a supported region

## Example Configuration

```json
{
  "stackName": "my-agent-stack",
  "region": "us-west-2",
  "agents": [
    {
      "name": "research",
      "containerImage": "123456789012.dkr.ecr.us-west-2.amazonaws.com/research-agent:latest",
      "memoryMB": 512,
      "timeoutSeconds": 300,
      "protocol": "HTTP"
    }
  ],
  "gateway": {
    "enabled": true,
    "name": "my_gateway",
    "protocol": "MCP"
  },
  "vpc": {
    "createVPC": true,
    "enableVPCEndpoints": true
  }
}
```

## Related Documentation

- [AWS AgentCore Developer Guide](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/)
- [AgentCore Runtime](agentcore.md) - Runtime code and server setup
- [agentkit-aws-cdk](https://github.com/agentplexus/agentkit-aws-cdk) - CDK constructs
- [agentkit-aws-pulumi](https://github.com/agentplexus/agentkit-aws-pulumi) - Pulumi components
