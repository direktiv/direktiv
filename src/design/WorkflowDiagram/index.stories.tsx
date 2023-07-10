import "../../AppLegacy.css";

import WorkflowDiagram from "./index";

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

export const WorkflowInvalid = () => (
  <div className="h-96">
    <WorkflowDiagram
      workflow="... /// invalid workflow"
      flow={["helloworld", "exit"]}
      instanceStatus="complete"
    />
  </div>
);
