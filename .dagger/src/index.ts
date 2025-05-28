import { Container, Directory, dag, func, object } from "@dagger.io/dagger";

@object()
export class HelloDagger {
  /**
   * Build the UI application
   */
  @func()
  async build(
    /**
     * Source directory containing the ui folder
     */
    source: Directory
  ): Promise<Directory> {
    // Get the ui subdirectory from source
    const uiDir = source.directory("ui");

    // Start with Node.js 20 slim image
    const container = dag
      .container()
      .from("node:20-slim")
      // Set up pnpm environment variables
      .withEnvVariable("PNPM_HOME", "/pnpm")
      .withEnvVariable("PATH", "/pnpm:$PATH", { expand: true })
      // Enable corepack and install pnpm
      .withExec(["corepack", "enable"])
      .withExec(["corepack", "prepare", "pnpm@9.15.4", "--activate"])
      // Set working directory
      .withWorkdir("/app")
      // Copy package files first for better caching
      .withFile("package.json", uiDir.file("package.json"))
      .withFile("pnpm-lock.yaml", uiDir.file("pnpm-lock.yaml"))
      // Install dependencies
      .withExec(["pnpm", "install", "--frozen-lockfile"])
      // Copy configuration files
      .withFile(".eslintrc.js", uiDir.file(".eslintrc.js"))
      .withFile(".prettierrc.mjs", uiDir.file(".prettierrc.mjs"))
      .withFile(".prettierignore", uiDir.file(".prettierignore"))
      .withFile("index.html", uiDir.file("index.html"))
      .withFile("postcss.config.cjs", uiDir.file("postcss.config.cjs"))
      .withFile("tailwind.config.cjs", uiDir.file("tailwind.config.cjs"))
      .withFile("tsconfig.json", uiDir.file("tsconfig.json"))
      .withFile("vite.config.mts", uiDir.file("vite.config.mts"))
      // Copy source directories
      .withDirectory("assets", uiDir.directory("assets"))
      .withDirectory("public", uiDir.directory("public"))
      .withDirectory("src", uiDir.directory("src"))
      .withDirectory("test", uiDir.directory("test"))
      // Run the build
      .withExec(["pnpm", "run", "build"]);

    // Return the build output directory (typically 'dist' for Vite projects)
    return container.directory("/app/dist");
  }
}
