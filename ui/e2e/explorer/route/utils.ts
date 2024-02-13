const methodsYaml = (methods: string[]) =>
  methods.map((method) => `\n  - "${method}"`).join("");

type Route = {
  path: string;
  timeout: number;
  methods: string[];
  plugins: Plugins;
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
  methods,
  plugins,
}: Route) => `direktiv_api: "endpoint/v1"
path: "${path}"
timeout: ${timeout}
methods:${methodsYaml(methods)}  
plugins:
  inbound: ${plugins.inbound ?? " []"}
  outbound:${plugins.outbound ?? " []"}
  auth:${plugins.auth ?? " []"}
  target:${plugins.target}`;
