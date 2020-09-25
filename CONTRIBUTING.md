# Contributing

Issues, whether bugs, tasks, or feature requests are essential for improving this project. We believe it should be as easy as possible to contribute changes that get things working in your environment. There are a few guidelines that we need contributors to follow so that we can keep on top of things.

## Code of Conduct

This project adheres to a [code of conduct](CODE_OF_CONDUCT.md). Please review this document before contributing to this project.

## Sign the CLA

Before you can contribute, you will need to sign the [Contributor License Agreement](https://cla-assistant.io/fairwindsops/controller-utils).

## Project Structure

controller-utils is a collection of packages that make it easier to perform common tasks in Kuberntes. Each package should be well tested.

## Getting Started

We label issues with the ["good first issue" tag](https://github.com/FairwindsOps/controller-utils/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) if we believe they'll be a good starting point for new contributors. If you're interested in working on an issue, please start a conversation on that issue, and we can help answer any questions as they come up. Another good place to start would be adding regression tests to existing plugins.

## Setting Up Your Development Environment
### Prerequisites
* A properly configured Golang environment with Go 1.13 or higher

## Running Tests

go test ./pkg/...

## Creating a New Issue

If you've encountered an issue that is not already reported, please create a [new issue](https://github.com/FairwindsOps/controller-utils/issues), choose `Bug Report`, `Feature Request` or `Misc.` and follow the instructions in the template. 


## Creating a Pull Request

Each new pull request should:

- Reference any related issues
- Pass existing tests and linting
- Contain a clear indication of if they're ready for review or a work in progress
- Be up to date and/or rebased on the master branch

## Creating a new release

* Update the `CHANGELOG.md` with any changes
* Create a new tag with the latest version number
