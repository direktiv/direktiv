import { Container, Directory, dag, func, object } from "@dagger.io/dagger";

@object()
export class Direktiv {
  /**
   * Build the UI application
   */
  @func()
  async buildUI(): Promise<Directory> {
    const uiDir = dag
      .directory("ui")
      .withoutDirectory("dist")
      .withoutDirectory("node_modules")
      .withoutDirectory("test-results")
      .withoutFile(".tsbuildinfo")
      .withoutFile(".eslintcache")
      .withoutFile("*.log")
      .withoutFile("**/.env*");

    const container = dag
      .container()
      .from("node:20.18.1-slim")
      .withEnvVariable("PNPM_HOME", "/pnpm")
      .withEnvVariable("PATH", "/pnpm:$PATH", { expand: true })
      .withExec(["corepack", "enable"])
      .withExec(["corepack", "prepare", "pnpm@9.15.4", "--activate"])
      .withWorkdir("/app")
      .withDirectory(".", uiDir)
      .withExec(["pnpm", "install", "--frozen-lockfile"])
      .withExec(["pnpm", "build"]);

    return container.directory("/app/dist");
  }
}
