<p align="center">
  <img src="assets/images/direktiv-logo-50.png" alt="direktiv">
</p>

<br>

<div align="center">

[![License](https://img.shields.io/badge/License-Apache--2.0-blue)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/direktiv/direktiv)](https://goreportcard.com/report/github.com/direktiv/direktiv) 
[![GitHub release](https://img.shields.io/github/release/direktiv/direktiv.svg)](https://github.com/direktiv/direktiv/releases/)
[![GitHub stars](https://badgen.net/github/stars/direktiv/direktiv)](https://github.com/direktiv/direktiv/stargazers/)
[![GitHub contributors](https://img.shields.io/github/contributors/direktiv/direktiv.svg)](https://github.com/direktiv/badges/graphs/contributors/)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/direktiv/direktiv/graphs/commit-activity)
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![Slack](https://img.shields.io/badge/Slack-Join%20Direktiv-4a154b?style=flat&logo=slack)](https://join.slack.com/t/direktiv-io/shared_invite/zt-zf7gmfaa-rYxxBiB9RpuRGMuIasNO~g)

</div>


<h1 align="center">Event-Driven Serverless Orchestration, Integration and Automation</h1>
<div align="center">
Run Workflows and Create Services in Seconds
</div>
</br>

# ⚠️ Project Archived — No Longer Maintained

**Status: End of Support**

This project has officially reached End of Life (EOL) and is no longer actively maintained or supported.

Effective **Oct-2025**:

- No further updates, bug fixes, or security patches will be provided.
- Issues and pull requests will no longer be reviewed or accepted.
- No technical support or maintenance commitments are provided.
- The repository remains publicly available for historical and reference purposes only.

**Security Notice:** This software should not be considered suitable for production deployment without an independent assessment of its security, compatibility, and ongoing maintenance requirements.

The source code remains available under its existing open-source licence.

We thank everyone who has contributed to or supported this project.

## Build a local development cluster

```bash
# Install task
brew install go-task

# Build local dev cluster for the first time
task default

# Rebuild existing cluster
task build:dev
