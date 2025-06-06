import {
  Container,
  Directory,
  Service,
  dag,
  func,
  object,
} from "@dagger.io/dagger";

@object()
export class Direktiv {
  /**
   * Build the UI application with pre built checks
   */
  @func()
  async buildUI(source: Directory): Promise<Directory> {
    const uiDir = source
      .directory("ui")
      .withoutDirectory("dist")
      .withoutDirectory("node_modules")
      .withoutDirectory("test-results")
      .withoutFile(".tsbuildinfo")
      .withoutFile(".eslintcache")
      .withoutFile("*.log")
      .withoutFile("**/.env*");

    const buildContainer = dag
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

    // Return the built "dist" directory
    return buildContainer.directory("/app/dist");
  }

  /**
   * Build and serve the UI application
   */
  @func()
  async serveUi(source: Directory): Promise<Service> {
    const builtApp = await this.buildUI(source);
    return dag
      .container()
      .from("nginx:alpine")
      .withWorkdir("/usr/share/nginx/html")
      .withDirectory(".", builtApp)
      .withExposedPort(80)
      .asService({ useEntrypoint: true });
  }
}
