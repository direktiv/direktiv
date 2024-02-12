const methodsYaml = (methods: string[]) =>
  methods.map((method) => `\n  - "${method}"`).join("");

type Route = {
  path: string;
  timeout: number;
  methods: string[];
};

export const createRouteYaml = ({
  path,
  timeout,
  methods,
}: Route) => `direktiv_api: "endpoint/v1"
path: "${path}"
timeout: ${timeout}
methods:${methodsYaml(methods)}  
plugins:
  inbound: []
  outbound: []
  auth: []
  target:
    type: "instant-response"
    configuration:
      status_code: 200`;
