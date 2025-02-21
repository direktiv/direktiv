type Route = {
  path: string;
  timeout: number;
  plugins: Plugins;
  allow_anonymous: boolean;
  methods: Record<string, object>;
};

type Plugins = {
  inbound?: string;
  outbound?: string;
  auth?: string;
  target: string;
};

export const normalizeText = (text: string): string =>
  text
    .split("\n")
    .filter((line) => line.trim().length > 0) // remove empty lines
    .join("\n")
    .trim();

export const createRouteYaml = ({
  path,
  timeout,
  plugins,
}: Route) => `x-direktiv-api: endpoint/v2
x-direktiv-config:
  allow_anonymous: true
  path: ${path}
  timeout: ${timeout}
  plugins:
    target:
      type: instant-response
      configuration:
        status_code: 200
    ${plugins.inbound ?? ""}
    ${plugins.outbound ?? ""}
    ${plugins.auth ?? ""}
get:
  responses:
    "200":
      description: ""
post:
  responses:
    "200":
      description: ""`;

export const removeLines = (
  text: string,
  lines: number,
  from: "top" | "bottom" = "top"
) => {
  const sliceArg = from === "top" ? [lines] : [0, -lines];
  return text
    .split("\n")
    .slice(...sliceArg)
    .join("\n");
};
