[![Slack](https://img.shields.io/badge/Slack-Join%20Direktiv-4a154b?style=flat&logo=slack)](https://join.slack.com/t/direktiv-io/shared_invite/zt-zf7gmfaa-rYxxBiB9RpuRGMuIasNO~g)

This repository contains the user interface for [direktiv](https://github.com/direktiv/direktiv).

<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/images/logo-dark.png">
    <source media="(prefers-color-scheme: light)" srcset="assets/images/logo-light.png">
    <img alt="direktiv logo" src="assets/images/logo-dark.png">
</picture>
</div>
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="assets/images/screenshot-dark.png">
  <source media="(prefers-color-scheme: light)" srcset="assets/images/screenshot-light.png">
  <img alt="direktiv ui" src="assets/images/screenshot-dark.png">
</picture>

# Requirements

- this repo requires a specific node version, to easily install the correct one, please install [nvm](https://github.com/nvm-sh/nvm)
- in the root directory, run `nvm use` to automatically switch to the required version
  - which will be read from `.nvmrc` file
  - please note that you always must run `nvm use` in every terminal session
  - when the required nvm version is not installed on your machine yet, you have to run nvm install once to do so
- this project uses [pnpm](https://pnpm.io/) as a package manager, make sure to install it via [`corepack`](https://pnpm.io/installation#using-corepack)

# Setup development environment

- run `pnpm` to install all dependencies
- create a `.env` file, copy the content from `.env.example` and change the variables to your needs
- run `pnpm dev` to start the dev server
- when you are using VSCode, make sure to install the recommended extensions for the best dev experience
  - VSCode should suggest you the extensions when you open this folder in VSCode, but you can also check them in the `.vscode/extensions.json` file

# Scripts you might want to run

- `pnpm run storybook` opens the storybook and documentation
- `pnpm run test` runs the tests in watch mode
- `pnpm run check` runs all the checks that will run in CI:
  - `pnpm run check:lint` runs the linter
  - `pnpm run check:types` checks all types
  - `pnpm run check:test` runs the tests in ci mode
  - `pnpm run check:prettier` checks if all prettier formatting was applied (`pnpm run prettier`)
  - `pnpm run check:deps` checks if all dependencies are up to date
  - `pnpm run check:knip` checks for unused code
- `pnpm run fix` fixes all autofixable issues:
  - `pnpm run fix:prettier` fixes all prettier formatting issues
  - `pnpm run fix:knip` fixes all issues related to unused code
- `pnpm run e2e:headless` runs all Playwright e2e tests in headless mode
- `pnpm run e2e:ui` runs all Playwright e2e tests in ui mode
- `pnpm run preview` locally preview the production build (requires you to run `pnpm build` first)

# Customize Logos

The direktiv logos are located in the folder `public/assets/logo` before the UI is built.

During the build process, the logos will be copied to `assets/logo` in the distribution folder, from where they will be served by the web server.

You can overwrite the files of the folder or mount a folder with your own logos under one of these paths, depending on whether you do it before or after the build process.

# Recommended Development Environment

## Visual Studio Code Users

If you're using [Visual Studio Code](https://code.visualstudio.com/) as your code editor, we highly recommend installing the extensions listed in the `.vscode/extensions.json` file. These extensions are designed to enhance your development experience by automatically resolving certain linting issues and formatting your code. When you open this project folder in VS Code for the first time, it should prompt you to install the recommended extensions.

## Non-VS Code Users

For those not using Visual Studio Code, we strongly recommend using [Prettier](https://prettier.io/) as your code formatter. Prettier will help you maintain a consistent code style, making it easier to collaborate with other developers. Prettier formatting checks will be enforced during the CI process.

# Code of Conduct

We have adopted the [Contributor Covenant](https://github.com/direktiv/.github/blob/master/CODE_OF_CONDUCT.md) code of conduct.

# Contributing

Any feedback and contributions are welcome. Read our [contributing guidelines](https://github.com/direktiv/.github/blob/master/CONTRIBUTING.md) for details.

# License

Distributed under the Apache 2.0 License. See `LICENSE` for more information.

# See Also

- The [direktiv.io](https://direktiv.io/) website.
- The direktiv [documentation](https://docs.direktiv.io/).
- The direktiv [blog](https://blog.direktiv.io/).
- The [Godoc](https://godoc.org/github.com/direktiv/direktiv) library documentation.
