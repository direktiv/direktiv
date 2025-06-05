import { Container, Directory, dag, func, object } from "@dagger.io/dagger";

@object()
export class Direktiv {
  /**
   * Build the UI application
   */
  @func()
  async buildUI(
    /**
     * Source directory containing the ui folder
     */
    source: Directory
  ): Promise<Directory> {
    // Get the ui subdirectory from source
    const uiDir = source
      .directory("ui")
      .withoutDirectory("dist")
      .withoutDirectory("node_modules")
      .withoutDirectory("test-results")
      .withoutFile(".tsbuildinfo")
      .withoutFile(".eslintcache")
      .withoutFile("*.log")
      .withoutFile("**/.env*");

    // Start with Node.js 20 slim image
    const container = dag
      .container()
      .from("node:20.18.1-slim")
      // Set up pnpm environment variables
      .withEnvVariable("PNPM_HOME", "/pnpm")
      .withEnvVariable("PATH", "/pnpm:$PATH", { expand: true })
      // Enable corepack and install pnpm
      .withExec(["corepack", "enable"])
      .withExec(["corepack", "prepare", "pnpm@9.15.4", "--activate"])
      // Set working directory
      .withWorkdir("/app")
      // Copy entire ui directory
      .withDirectory(".", uiDir)
      // Install dependencies
      .withExec(["pnpm", "install", "--frozen-lockfile"])
      // Run the build
      .withExec(["pnpm", "build"]);

    // Return the build output directory (typically 'dist' for Vite projects)
    return container.directory("/app/dist");
  }
}
