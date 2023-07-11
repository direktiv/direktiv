import "../../AppLegacy.css";

import Button from "../Button";
import { ButtonBar } from "../ButtonBar";
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
  const [workflow, setWorkflow] = useState(exampleWorkflow2);

  return (
    <div className="flex h-96 flex-col gap-y-5">
      <ButtonBar className="">
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
      </ButtonBar>
      <WorkflowDiagram
        workflow={workflow}
        flow={["helloworld", "exit"]}
        instanceStatus="complete"
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
