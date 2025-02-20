type Route = {
  path: string;
  timeout: number;
  methods: string[];
  plugins: Plugins;
  allow_anonymous: boolean;
  patch: {
    responses: {
      "200": {
        description: string;
      };
    };
  };
};

type Plugins = {
  inbound?: string;
  outbound?: string;
  auth?: string;
  target: string;
};

export const createRouteYaml = ({
  path,
  timeout,
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
