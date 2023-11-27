import Button from "../Button";
import { ButtonBar } from "../ButtonBar";
import { Card } from "../Card";
import { Orientation } from "./types";
import WorkflowDiagram from "./index";
import { useState } from "react";

const exampleWorkflow = `description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transition: exit
  transform:
    result: Hello world!
- id: exit
  type: noop
`;

const exampleWorkflow2 = `description: A simple 'error' state workflow that checks an email attempts to validate it.
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
    result: "Email is valid."`;

const ComplexWorkflow = `description: runs a text prompt against the stable diffusion model via replicate.com API (https://replicate.com/docs/reference/http)
functions:
  - id: request
    image: direktiv/request:v4
    type: knative-workflow
# we need a prompt string and num_outputs integer to determine how much images we want to generate
states:
  - id: validate-input-2
    type: validate
    schema:
      title: Prompt
      type: object
      properties:
        prompt:
          type: string
          title: Prompt
          minLength: 10
          maxLength: 2000
        num_outputs:
          type: number
          title: "number of images"
          default: 2
          enum: 
            - 1
            - 2
            - 3
            - 4
    catch:
      - error: direktiv.schema.*
        transition: prompt-invalid
    transition: request-prediction
  - id: prompt-invalid
    type: noop
    transform:
      result: "Prompt is not valid"
  # Let's create a prediction via POST request
  - id: request-prediction
    transform: 'jq({predictionUrl: "https://api.replicate.com/v1/predictions/(.return.body.id)", error: .return.body.error, statusCode: .return."status-code"})'
    type: action
    action:
      secrets: ["REPLICATE_API_KEY"]
      function: request
      input:
        method: "POST"
        headers:
          "Content-Type": "application/json"
          "Authorization": "Token jq(.secrets.REPLICATE_API_KEY)"
        body:
          version: "f178fa7a1ae43a9a9af01b833b9d2ecf97b1bcb0acfd2dc5dd04895e042863f1"
          input:
            prompt: jq(.prompt)
            num_outputs: jq(.num_outputs)
        url: "https://api.replicate.com/v1/predictions"
    transition: evaluate-created-prediction
  # check if the prediction was created
  - id: evaluate-created-prediction
    type: switch
    defaultTransition: sleep-before-request
    conditions:
      - condition: "jq(.error != null)"
        transition: api-response-error
      # this will hit, when the test phase has expired and you need to enter a credit card
      - condition: "jq(.statusCode == 402)"
        transition: api-response-error
  # handle api errors
  - id: api-response-error
    type: noop
    transform:
      result: "The API responded with an error. jq(.)"
  # the image generation needs some time (on average around 14 seconds)
  - id: sleep-before-request
    type: delay
    duration: PT8S
    transition: request-image
  # get the state of the prediction
  - id: request-image
    type: action
    transform: "jq({predictionUrl: .predictionUrl, error: .return.body.error, error: .return.body.error, images: .return.body.output})"
    transition: evaluate-image-response
    action:
      secrets: ["REPLICATE_API_KEY"]
      function: request
      input:
        method: "GET"
        headers:
          "Content-Type": "application/json"
          "Authorization": "Token jq(.secrets.REPLICATE_API_KEY)"
        url: "jq(.predictionUrl)"
  # the response could have an error, or the image might not be ready yet
  - id: evaluate-image-response
    type: switch
    defaultTransition: done
    conditions: 
      - condition: "jq(.error != null)"
        transition: api-response-error
      - condition: "jq(.images == null)"
        transition: done
  - id: done
    type: noop
    transform:
      images: "jq(.images)"`;

export default {
  title: "Components/WorkflowDiagram",
  component: WorkflowDiagram,
  argTypes: {
    instanceStatus: {
      options: ["complete", "failed", "pending"],
      control: { type: "select" },
      defaultValue: "pending",
      description:
        "Status of current instance. This is used to display if flow is complete with animated connections.",
      table: {
        type: { summary: "string" },
      },
    },
    workflow: {
      description: "YAML string of workflow.",
      table: {
        type: { summary: "string" },
      },
    },
    flow: {
      description:
        "Array of executed states in an instance. Example - ['noopA', 'noopB']",
      table: {
        type: { summary: "string[]" },
      },
    },
    disabled: {
      description: "Disables diagram zoom-in",
      table: {
        type: { summary: "boolean" },
      },
    },
  },
};

export const Workflow = () => (
  <div className="h-96">
    <WorkflowDiagram workflow={exampleWorkflow} />
  </div>
);

export const WorkflowInstancePending = () => (
  <div className="h-96">
    <WorkflowDiagram
      workflow={exampleWorkflow}
      flow={["helloworld"]}
      instanceStatus="pending"
    />
  </div>
);

export const WorkflowInstanceComplete = () => (
  <div className="h-96">
    <WorkflowDiagram
      workflow={exampleWorkflow}
      flow={["helloworld", "exit"]}
      instanceStatus="complete"
    />
  </div>
);

export const UpdateWorkflow = () => {
  const [workflow, setWorkflow] = useState(exampleWorkflow);
  const [orientation, setOrientation] = useState<Orientation>("horizontal");

  return (
    <Card className="flex h-96 flex-col gap-y-5 p-5" background="weight-1">
      <ButtonBar>
        <Button
          onClick={() => {
            setWorkflow(exampleWorkflow);
          }}
        >
          Example 1
        </Button>
        <Button
          onClick={() => {
            setWorkflow(exampleWorkflow2);
          }}
        >
          Example 2
        </Button>
        <Button
          onClick={() => {
            setWorkflow("");
          }}
        >
          Empty Workflow
        </Button>
        <Button
          variant="primary"
          onClick={() => {
            setOrientation((old) =>
              old === "horizontal" ? "vertical" : "horizontal"
            );
          }}
        >
          Change Orientation
        </Button>
      </ButtonBar>
      <WorkflowDiagram
        workflow={workflow}
        flow={["helloworld", "exit"]}
        instanceStatus="complete"
        orientation={orientation}
      />

      <div>
        <pre>{JSON.stringify(workflow)}</pre>
      </div>
    </Card>
  );
};

export const ComplexWorkflowDiagram = () => {
  const allStates = [
    "validate-input-2",
    "request-prediction",
    "evaluate-created-prediction",
    "sleep-before-request",
    "request-image",
    "evaluate-image-response",
    "done",
  ];

  const [progress, setProgress] = useState(0);
  const [orientation, setOrientation] = useState<Orientation>("horizontal");

  return (
    <div className="flex h-[600px] flex-col gap-y-5">
      <ButtonBar>
        <Button
          onClick={() => {
            setProgress((old) => {
              if (old === allStates.length) {
                return 0;
              }
              return old + 1;
            });
          }}
        >
          Simulate Progress
        </Button>
        <Button
          onClick={() => {
            setOrientation((old) =>
              old === "horizontal" ? "vertical" : "horizontal"
            );
          }}
        >
          Change Orientation
        </Button>
      </ButtonBar>
      <WorkflowDiagram
        workflow={ComplexWorkflow}
        flow={allStates.slice(0, progress)}
        instanceStatus="complete"
        orientation={orientation}
      />
    </div>
  );
};

export const WorkflowInvalid = () => (
  <div className="h-96">
    <WorkflowDiagram
      workflow="... /// invalid workflow"
      flow={["helloworld", "exit"]}
      instanceStatus="complete"
    />
  </div>
);
