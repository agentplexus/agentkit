# Roadmap

This document outlines the development roadmap for agentkit and its companion modules.

## Completed (v0.3.0)

- [x] OmniVault integration for unified secret management
- [x] Config file loading from JSON/YAML (`config/file.go`)
- [x] `Load(ctx, LoadOptions)` unified config loading function
- [x] Auto-detection of AWS environment (ECS, Lambda, EC2)
- [x] Protocol configuration for agents (HTTP, MCP, A2A)
- [x] Authorization and gateway configuration

## Completed (v0.2.0)

- [x] IaC configuration package (`platforms/agentcore/iac/`)
- [x] AWS CDK constructs ([agentkit-aws-cdk](https://github.com/agentplexus/agentkit-aws-cdk))
- [x] Pulumi components ([agentkit-aws-pulumi](https://github.com/agentplexus/agentkit-aws-pulumi))
- [x] Pure CloudFormation generation
- [x] AWS deployment guide documentation

## In Progress

- [ ] Terraform modules ([agentkit-terraform](https://github.com/agentplexus/agentkit-terraform))

## Planned

### Deployment Targets

- [ ] ECS/Fargate support
- [ ] Extended Helm chart library for Kubernetes

### Runtime Features

- [ ] Streaming response support
- [ ] Additional LLM provider adapters
- [ ] Enhanced observability integrations

### Multi-Cloud

- [ ] GCP Pulumi components (agentkit-gcp-pulumi)
- [ ] Azure Pulumi components (agentkit-azure-pulumi)
- [ ] GCP Terraform modules
- [ ] Azure Terraform modules

## Module Overview

| Module | Status | Purpose |
|--------|--------|---------|
| [agentkit](https://github.com/agentplexus/agentkit) | GA | Core library, shared IaC config, CloudFormation |
| [agentkit-aws-cdk](https://github.com/agentplexus/agentkit-aws-cdk) | GA | AWS CDK constructs |
| [agentkit-aws-pulumi](https://github.com/agentplexus/agentkit-aws-pulumi) | GA | AWS Pulumi components |
| [agentkit-terraform](https://github.com/agentplexus/agentkit-terraform) | Planned | Terraform modules (AWS/GCP/Azure) |

## Contributing

Contributions are welcome! If you're interested in working on any roadmap item:

1. Check existing issues for related discussions
2. Open an issue to discuss your approach
3. Submit a pull request

## Versioning

- **v0.1.0** - Initial release with AgentCore runtime support
- **v0.2.0** - IaC support (CDK, Pulumi, CloudFormation)
- **v0.3.0** - OmniVault integration, config file loading
- **v0.4.0** - Terraform modules (planned)
- **v0.5.0** - ECS/Fargate support (planned)
