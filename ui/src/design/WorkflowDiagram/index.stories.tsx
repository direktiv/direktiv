import Button from "../Button";
import { ButtonBar } from "../ButtonBar";
import { Card } from "../Card";
import { InstanceFlowSchemaType } from "~/api/instances/schema";
import { Orientation } from "./types";
import WorkflowDiagram from "./index";

import { useState } from "react";

const exampleWorkflowInitial = {
  flow: [],
  states: [
    {
      name: "hello-world",
      start: true,
      finish: false,
      visited: false,
      failed: false,
      transitions: ["exit"],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
    {
      name: "exit",
      start: false,
      finish: true,
      visited: false,
      failed: false,
      transitions: [],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
  ],
} satisfies InstanceFlowSchemaType;

const exampleWorkflowStepOne = {
  flow: ["hello-world"],
  states: [
    {
      name: "hello-world",
      start: true,
      finish: false,
      visited: true,
      failed: false,
      transitions: ["exit"],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
    {
      name: "exit",
      start: false,
      finish: true,
      visited: false,
      failed: false,
      transitions: [],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
  ],
} satisfies InstanceFlowSchemaType;

const exampleWorkflowStepTwo = {
  flow: ["hello-world", "exit"],
  states: [
    {
      name: "hello-world",
      start: true,
      finish: false,
      visited: true,
      failed: false,
      transitions: ["exit"],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
    {
      name: "exit",
      start: false,
      finish: true,
      visited: true,
      failed: false,
      transitions: [],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
  ],
} satisfies InstanceFlowSchemaType;

const exampleWorkflowFailed = {
  flow: ["stateFirst", "stateSecond"],
  states: [
    {
      name: "stateFirst",
      start: true,
      finish: false,
      visited: true,
      failed: false,
      transitions: ["stateSecond"],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
    {
      name: "stateSecond",
      start: false,
      finish: false,
      visited: true,
      failed: true,
      transitions: ["stateThird"],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
    {
      name: "stateThird",
      start: false,
      finish: true,
      visited: false,
      failed: false,
      transitions: [],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
  ],
} satisfies InstanceFlowSchemaType;

const exampleWorkflowComplete = {
  flow: ["hello-world", "exit"],
  states: [
    {
      name: "hello-world",
      start: true,
      finish: false,
      visited: true,
      failed: false,
      transitions: ["exit"],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
    {
      name: "exit",
      start: false,
      finish: true,
      visited: true,
      failed: false,
      transitions: [],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
  ],
} satisfies InstanceFlowSchemaType;

const exampleWorkflowComplex = {
  flow: ["hello-world", "actionTwo", "exit"],
  states: [
    {
      name: "hello-world",
      start: true,
      finish: false,
      visited: true,
      failed: false,
      transitions: ["actionOne", "actionTwo"],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
    {
      name: "actionOne",
      start: false,
      finish: false,
      visited: false,
      failed: false,
      transitions: ["exit"],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
    {
      name: "actionTwo",
      start: false,
      finish: false,
      visited: true,
      failed: false,
      transitions: ["exit"],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
    {
      name: "exit",
      start: false,
      finish: true,
      visited: true,
      failed: false,
      transitions: [],
      events: [],
      conditions: [],
      catch: [],
      transition: "",
      defaultTransition: "",
    },
  ],
} satisfies InstanceFlowSchemaType;

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
      description: "JSON of workflow.",
      table: {
        type: { summary: "object" },
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
    <WorkflowDiagram data={exampleWorkflowInitial} />
  </div>
);

export const WorkflowInstancePending = () => (
  <div className="h-96">
    <WorkflowDiagram data={exampleWorkflowStepTwo} instanceStatus="pending" />
  </div>
);

export const WorkflowInstanceFailed = () => (
  <div className="h-96">
    <WorkflowDiagram data={exampleWorkflowFailed} instanceStatus="failed" />
  </div>
);

export const WorkflowInstanceComplete = () => (
  <div className="h-96">
    <WorkflowDiagram data={exampleWorkflowComplete} instanceStatus="complete" />
  </div>
);

export const UpdateWorkflow = () => {
  const [workflow, setWorkflow] = useState<InstanceFlowSchemaType>(
    exampleWorkflowInitial
  );
  const [status, setStatus] = useState<"pending" | "complete">("pending");
  const [orientation, setOrientation] = useState<Orientation>("horizontal");

  return (
    <Card className="flex h-96 flex-col gap-y-5 p-5" background="weight-1">
      <ButtonBar>
        <Button
          onClick={() => {
            setWorkflow(exampleWorkflowInitial);
            setStatus("pending");
          }}
        >
          00 - Workflow loaded
        </Button>
        <Button
          onClick={() => {
            setWorkflow(exampleWorkflowStepOne);
            setStatus("pending");
          }}
        >
          01 - Hello-World
        </Button>
        <Button
          onClick={() => {
            setWorkflow(exampleWorkflowStepTwo);
            setStatus("pending");
          }}
        >
          02 - Exit
        </Button>

        <Button
          onClick={() => {
            setWorkflow(exampleWorkflowStepTwo);
            setStatus("complete");
          }}
        >
          03 - Workflow complete
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
        data={workflow}
        instanceStatus={status}
        orientation={orientation}
      />

      <div>
        <pre>{JSON.stringify(workflow)}</pre>
      </div>
    </Card>
  );
};

export const ComplexWorkflowDiagram = () => (
  <div className="h-[600px]">
    <WorkflowDiagram data={exampleWorkflowComplex} instanceStatus="complete" />
  </div>
);
