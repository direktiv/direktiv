export const workflowWithSecrets = `// Simple example workflow to test secrets
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateHello",
};

function stateHello(input): StateFunction<unknown> {
  const secrets = getSecrets(["one", "two"]);

  const result = Object.entries(secrets).map(([, secret]) => secret);

  return finish({ secrets: result });
}
`;
