# controller-utils

[![FairwindsOps](https://circleci.com/gh/FairwindsOps/controller-utils.svg?style=svg)](https://circleci.com/gh/FairwindsOps/controller-utils)
[![Apache 2.0 license](https://img.shields.io/badge/license-Apache2-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0)

This is a library of Go functions to assist in building Kubernetes Controllers.

The `pkg/controller` package contains the main functionality. `pkg/log` contains helpers around logging. You can pass in a [logr](https://github.com/go-logr/logr) object to control the logs of this library.

## Join the Fairwinds Open Source Community

The goal of the Fairwinds Community is to exchange ideas, influence the open source roadmap, and network with fellow Kubernetes users. [Chat with us on Slack](https://join.slack.com/t/fairwindscommunity/shared_invite/zt-e3c6vj4l-3lIH6dvKqzWII5fSSFDi1g) or [join the user group](https://www.fairwinds.com/open-source-software-user-group) to get involved!

## Contributing

PRs welcome! Check out the [Contributing Guidelines](CONTRIBUTING.md) and
[Code of Conduct](CODE_OF_CONDUCT.md) for more information.


## Other Projects from Fairwinds

Enjoying controller-utils? Check out some of our other projects:
* [Polaris](https://github.com/FairwindsOps/Polaris) - Audit, enforce, and build policies for Kubernetes resources, including over 20 built-in checks for best practices
* [Goldilocks](https://github.com/FairwindsOps/Goldilocks) - Right-size your Kubernetes Deployments by compare your memory and CPU settings against actual usage
* [Pluto](https://github.com/FairwindsOps/Pluto) - Detect Kubernetes resources that have been deprecated or removed in future versions
* [Nova](https://github.com/FairwindsOps/Nova) - Check to see if any of your Helm charts have updates available
* [rbac-manager](https://github.com/FairwindsOps/rbac-manager) - Simplify the management of RBAC in your Kubernetes clusters

Or [check out the full list](https://www.fairwinds.com/open-source-software?utm_source=controller-utils&utm_medium=controller-utils&utm_campaign=controller-utils)
