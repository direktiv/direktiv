export const bashServiceSrc = `{
  "image": "direktiv/bash:dev",
  "scale": 1,
  "size": "small"
}`

export const bashServiceScale3Src = `{
  "image": "direktiv/bash:dev",
  "scale": 3,
  "size": "small"
}`

export const bashServiceScale0Src = `{
  "image": "direktiv/bash:dev",
  "scale": 0,
  "size": "small"
}`

export const echoServiceSrc = `{
  "image": "direktiv/echo",
  "scale": 1,
  "size": "small"
}`

export const bashServiceWithEnvsSrc = `{
  "image": "direktiv/bash:dev",
  "scale": 1,
  "size": "small",
  "envs": [
    { "name": "FOO1", "value": "bar1" },
    { "name": "FOO2", "value": "bar2" }
  ]
}`

export const workflowUsingServiceSrc = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst() {
  var payload = {
    commands: [
      {
        command: "echo direktiv-bash-service-ok",
      },
    ],
  };

  const serviceResponse = execService({
    scope: "namespace",
    path: "/service.svc.json",
    payload: payload,
    retries: 3,
  });

  return finish(serviceResponse);
}
`

export const workflowUsingSystemServiceSrc = `
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst() {
  var payload = {
    commands: [
      {
        command: "echo direktiv-bash-service-ok",
      },
    ],
  };

  const serviceResponse = execService({
    scope: "system",
    path: "/system-service.svc.json",
    payload: payload,
    retries: 3,
  });

  return finish(serviceResponse);
}
`

export const workflowFilesInSystemServiceSrc = `const SYSTEM_SERVICE_PATH = "/bash.svc.json";

const flow: FlowDefinition = {
  type: "default",
  timeout: "PT5M",
  state: "stateRunService",
};

function stateRunService() {
  const result = execService({
    scope: "system",
    path: SYSTEM_SERVICE_PATH,
    payload: {
      commands: [
        {
          command: "ls -la",
        },
        {
          command: "cat foobar",
        },
        {
          command: "cat test.wf.ts",
        },
      ],
    },
    files: [
      { name: "foobar", scope: "namespace", permission: "0755" },
      { name: "/test.wf.ts", scope: "file" },
    ],
  });

  return finish(result);
}
`
