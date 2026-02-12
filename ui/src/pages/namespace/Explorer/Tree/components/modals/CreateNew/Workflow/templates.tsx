export const hello = {
  name: "hello",
  data: `const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst(): StateFunction<unknown> {
  return finish({ data: "hello world" })  
}
`,
};

const input = {
  name: "input",
  data: `// a simple workflow example demonstrating flow from one state
// to the next, failing with an error, and typescript evaluation
// of the input data. Expects an input payload like
// { "data": "foo" } that it will evaluate, or fail.

// workflow must include a flow definition
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

// you can define variables
const error = 'input must contain { "data": "string" or number }';

// a regular typescript function that you can reference elsewhere
function formatMessage(data: string | number, type: string) {
  return { message: \`\${data} is a \${type}\` };
}

// define states that the workflow will progress through
function stateFirst(input): StateFunction<unknown> {
  const { data } = input;
  if (!data) {
    // this will fail the workflow
    throw Error(error);
  }
  // a state must return a transition() or finish()
  return transition(stateSecond, data);
}

function stateSecond(data): StateFunction<unknown> {
  const type = typeof data;
  if (type === "string" || type === "number") {
    const message = formatMessage(data, type);
    return finish(message);
  }
  // a state must return a transition() or finish()
  return finish({ error });
}
`,
};

const actions = {
  name: "actions",
  data: `const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

const d = generateAction({
  image: "ubuntu:24.04",
  cmd: "/usr/share/direktiv/direktiv-cmd",
  size: "small",
  envs: [
    {
      name: "myenv",
      value: "myenvvalue",
    },
  ],
});

function stateFirst(): StateFunction<unknown> {
  var payload = {
    commands: [
      {
        command: "ls -la",
      },
    ],
  };
  let result = d(payload);
  return finish(result);
}
`,
};

const services = {
  name: "services",
  data: `const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst(input) {
  var payload = {
    commands: [
      {
        command: "ls -la",
      },
    ],
  };

  execService({
    scope: "namespace",
    path: "/myservice.svc.json",
    payload,
  });

  execService({
    scope: "system",
    path: "/system-service.svc.json",
    payload: payload,
    retries: 3,
  });

  return finish("done");
}
`,
};

const secrets = {
  name: "secrets",
  data: `const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateVerifySecrets",
};

function stateVerifySecrets(): StateFunction<unknown> {
  const secrets = getSecrets(["one", "two"]);

  Object.values(secrets).forEach((value) => {
    if (value.length === 0) {
      throw Error("The secrets must not be empty");
    }
  });

  if (secrets.one !== secrets.two) {
    throw Error("The secrets must match");
  }

  return finish(secrets);
}
`,
};

const variables = {
  name: "variables",
  data: `const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst(): StateFunction<unknown> {
  return finish("TBD");
}
`,
};

const error = {
  name: "hello",
  data: `const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

function stateFirst(): StateFunction<unknown> {
  if (true) {
    throw Error("this was set up to fail");
  }

  return finish("unreachable");
}
`,
};

const branches = {
  name: "branches",
  data: `// this will randomly progress through alternative states,
// resulting in an interesting flow visualization
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateA",
};

function stateA(): StateFunction<unknown> {
  if (Math.random() > 0.25) {
    return transition(stateB, {});
  }
  return transition(stateC, {});
}

function stateB(): StateFunction<unknown> {
  if (Math.random() > 0.25) {
    return transition(stateD, {});
  }
  return finish({ message: "finishing from stateB" });
}

function stateC(): StateFunction<unknown> {
  return finish({ message: "finishing from stateC" });
}

function stateD(): StateFunction<unknown> {
  if (Math.random() > 0.5) {
    return transition(stateE, {});
  }
  return transition(stateC, {});
}

function stateE(input): StateFunction<unknown> {
  return finish({ message: "finishing from stateE" });
}
`,
};

export const consumeEvent = {
  name: "consumeEvent",
  data: `direktiv_api: workflow/v1
functions:
- id: greeter
  image: direktiv/greeting:v3
  type: knative-workflow
description: A simple 'consumeEvent' state that listens for the greetingcloudevent generated from the template 'generate-event'.
states:
- id: ce
  type: consumeEvent
  event:
    type: greetingcloudevent
  timeout: PT1H
  transition: greet
- id: greet
  type: action
  action:
    function: greeter
    input: jq(.greetingcloudevent.data)
  transform:
    greeting: jq(.return.greeting)
`,
};

const templates = [
  hello,
  input,
  actions,
  services,
  secrets,
  variables,
  error,
  branches,
] as const;

export default templates;
