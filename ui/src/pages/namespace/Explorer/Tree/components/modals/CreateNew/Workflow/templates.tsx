const tsdemo = {
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

// Todo: examples below (in yaml) are no longer valid. Replace them with
// ts versions (or delete if not needed)

export const noop = {
  name: "noop",
  data: `direktiv_api: workflow/v1
description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`,
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const action = {
  name: "action",
  data: `direktiv_api: workflow/v1
description: A simple 'action' state that sends a get request
functions:
- id: get
  image: direktiv/request:v4
  type: knative-workflow
states:
- id: getter 
  type: action
  action:
    function: get
    input: 
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/todos/1"
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

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const delay = {
  name: "delay",
  data: `direktiv_api: workflow/v1
description: A simple 'delay' state that waits for 5 seconds
states:
- id: delay
  type: delay
  duration: PT5S
`,
};

export const error = {
  name: "error",
  data: `direktiv_api: workflow/v1
description: A simple 'error' state workflow that checks an email attempts to validate it.
states:
- id: data
  type: noop
  transform: 
    email: "trent.hilliamdirektiv.io"
  transition: validate-email
- id: validate-email
  type: validate
  subject: jq(.)
  schema:
    type: object
    properties:
      email:
        type: string
        format: email
  catch:
  - error: direktiv.schema.*
    transition: email-not-valid 
  transition: email-valid
- id: email-not-valid
  type: error
  error: direktiv.schema.*
  message: "email '.email' is not valid"
- id: email-valid
  type: noop
  transform: 
    result: "Email is valid."
`,
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const foreach = {
  name: "foreach",
  data: `direktiv_api: workflow/v1
description: A simple 'foreach' state that solves expressions
functions: 
- id: solve
  image: direktiv/solve:v3
  type: knative-workflow
states:
- id: data
  type: noop
  transform: 
    expressions: ["4+10", "15-14", "100*3","200/2"] 
  transition: solve
- id: solve
  type: foreach
  array: 'jq([.expressions[] | {expression: .}])'
  action:
    function: solve
    input:
      x: jq(.expression)
  transform:
    solved: jq(.return)
`,
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const generateEvent = {
  name: "generateEvent",
  data: `direktiv_api: workflow/v1
description: A simple 'generateEvent' state that sends data to a greeting listener.
states:
- id: generate
  type: generateEvent
  event:
    type: greetingcloudevent
    source: Direktiv
    data: 
      name: "Trent"
`,
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const generateSolveEvent = {
  name: "generateSolveEvent",
  data: `direktiv_api: workflow/v1
description: A simple 'generateEvent' state that sends an expression to a solve listener.
states:
- id: generate
  type: generateEvent
  event:
    type: solveexpressioncloudevent
    source: Direktiv
    data: 
      x: "10+5"
`,
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const getAndSet = {
  name: "getAndSet",
  data: `direktiv_api: workflow/v1
description: "Simple Counter getter and setter variable example"
states:
  - id: counter-get
    type: getter 
    transition: counter-set
    variables:
    - key: ExampleVariableCounter
      scope: workflow
    transform: 'jq(. += {"newCounter": (.var.ExampleVariableCounter + 1)})'
  - id: counter-set
    type: setter
    variables:
      - key: ExampleVariableCounter
        scope: workflow 
        value: 'jq(.newCounter)'
`,
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const parallel = {
  name: "parallel",
  data: `direktiv_api: workflow/v1
description: A simple 'parallel' state workflow that runs solve container to solve expressions.
functions: 
- id: solve
  image: direktiv/solve:v3
  type: knative-workflow
states:
- id: run
  type: parallel
  actions:
  - function: solve
    input: 
      x: "10*2"
  - function: solve
    input:
      x: "10%2"
  - function: solve
    input:
      x: "10-2"
  - function: solve
    input:
      x: "10+2"
  # Mode 'and' waits for all actions to be completed
  # Mode 'or' waits for the first action to be completed
  mode: and
`,
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const validate = {
  name: "validate",
  data: `direktiv_api: workflow/v1
description: A simple 'validate' state workflow that checks an email
states:
- id: data
  type: noop
  transform:
    email: "trent.hilliam@direktiv.io"
  transition: validate-email
- id: validate-email
  type: validate
  subject: jq(.)
  schema:
    type: object
    properties:
      email:
        type: string
        format: email
  catch:
  - error: direktiv.schema.*
    transition: email-not-valid 
  transition: email-valid
- id: email-not-valid
  type: noop
  transform:
    result: "Email is not valid."
- id: email-valid
  type: noop
  transform:
    result: "Email is valid."
`,
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const switchState = {
  name: "switch",
  data: `direktiv_api: workflow/v1
description: A simple 'switch' state that checks whether the age provided is older than 18.
states:
- id: data
  type: noop
  transform:
    age: 27
  transition: check
- id: check
  type: switch
  conditions:
  - condition: 'jq(.age > 18)'
    transition: olderthan18
  defaultTransition: youngerthan18
- id: olderthan18
  type: noop
  transform: 
    result: "You are older than 18."
- id: youngerthan18
  type: noop
  transform: 
    result: "You are younger than 18."
`,
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const eventXor = {
  name: "eventXor",
  data: `direktiv_api: workflow/v1
functions:
- id: greeter
  image: direktiv/greeting:v3
  type: knative-workflow
- id: solve2
  image: direktiv/solve:v3
  type: knative-workflow
description: A simple 'eventXor' that waits for events to be received.
states:
- id: event-xor
  type: eventXor
  timeout: PT1H
  events:
  - event: 
      type: solveexpressioncloudevent
    transition: solve
  - event: 
      type: greetingcloudevent
    transition: greet
- id: greet
  type: action
  action:
    function: greeter
    input: jq(.greetingcloudevent.data)
  transform: 
    greeting: jq(.return.greeting)
- id: solve
  type: action
  action:
    function: solve2
    input: jq(.solveexpressioncloudevent.data)
  transform: 
    solvedexpression: jq(.return)
`,
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const eventAnd = {
  name: "eventAnd",
  data: `direktiv_api: workflow/v1
functions:
- id: greeter
  image: direktiv/greeting:v3
  type: knative-workflow
- id: solve
  image: direktiv/solve:v3
  type: knative-workflow
description: A simple 'eventAnd' that waits for events to be received.
states:
- id: event-and
  type: eventsAnd
  timeout: PT1H
  events:
  - type: greetingcloudevent
  - type: solveexpressioncloudevent
  transition: greet
- id: greet
  type: action
  action:
    function: greeter
    input: jq(.greetingcloudevent.data)
  transform: 
    greeting: jq(.return.greeting)
    ceevent: jq(.solveexpressioncloudevent)
  transition: solve
- id: solve
  type: action
  action:
    function: solve
    input: jq(.ceevent.data)
  transform: 
    msggreeting: jq(.greeting)
    solvedexpression: jq(.return)
`,
};

const templates = [tsdemo] as const;

export default templates;
