export const tsdemo = {
  name: "tsdemo",
  data: `// Simple example workflow that validates type of input
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT30S",
  state: "stateFirst",
};

const error = 'input must contain { "data": "string" or number }'

function stateFirst(input): StateFunction<unknown> {
  const { data } = input;
  if (!data) {
    return finish({ error });
  }
  return transition(stateSecond, data);
}

function stateSecond(data): StateFunction<unknown> {
  const type = typeof data;
  if (type === "string" || type === "number") {
    const message = formatMessage(data, type);
    return finish(message)
  }
  return finish({ error })
}

function formatMessage(data: string | number, type: string) {
  return { message: \`\${data} is a \${type}\` }
};
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

const templates = [tsdemo] as const;

export default templates;
