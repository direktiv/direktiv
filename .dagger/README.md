# Dagger

## Install Dagger CLI

```bash
brew install dagger/tap/dagger
```

## General Dagger commands

run from the root of the project

- List available functions: `dagger functions`
- Call a function: `dagger call <FUNCTION>`

## Direktiv Dagger functions

- Build the ui: `dagger call build-ui --source=.`
- Build and serve the UI locally `dagger call serve-ui --source=. up --ports=8080:80`
