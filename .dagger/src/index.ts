import { Container, Directory, dag, func, object } from "@dagger.io/dagger";

@object()
export class Direktiv {
  /**
   * Build the UI application
   */
  @func()
  async buildUI(
    /**
     * source directory
     */
    source: Directory
  ): Promise<Directory> {
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
   * Run the UI application in a container
   */
  @func()
  async nginx(
    /**
     * server directory
     */
    builtUI: Directory
  ): Promise<Container> {
    const runtimeContainer = dag
      .container()
      .from("nginx:alpine") // Use Nginx to serve the built files
      .withWorkdir("/usr/share/nginx/html")
      .withDirectory(".", builtUI) // Copy the built files to the Nginx container
      .withExposedPort(8080) // Expose port 8080
      .withExec(["nginx", "-g", "daemon off;"]); // Start the Nginx server

    // Return the container
    return runtimeContainer;
  }

  /**
   * serve UI
   */
  @func()
  async serveUi(source: Directory): Promise<void> {
    const builtApp = await this.buildUI(source);
    const container = await this.nginx(builtApp);

    // Log the exposed port for debugging or access
    console.log("Server is running and exposed on port 8080");
  }
}
