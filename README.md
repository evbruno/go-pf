# go-pf

![build](https://github.com/evbruno/go-pf/actions/workflows/build.yml/badge.svg)
[![Build and Release](https://github.com/evbruno/go-pf/actions/workflows/release.yml/badge.svg)](https://github.com/evbruno/go-pf/actions/workflows/release.yml)

![go-pf logo|500](./go-pf.jpeg)

A simple command-line tool to find duplicated port usage across Kubernetes services within a specified namespace and context, and generate `kubectl port-forward` commands.

## Features

*   Lists all services in a given namespace and context along with their ports.
*   Identifies and reports duplicated port usage across services.
*   Generates `kubectl port-forward` commands for easy port forwarding.

## Installation

```bash
go build .
./go-pf <k8s-context> <k8s-namespace>
````
