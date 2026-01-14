import Button from "../Button";
import { ButtonBar } from "../ButtonBar";
import { Card } from "../Card";
import { Orientation } from "./types";
import WorkflowDiagram from "./index";
import { useState } from "react";
import { Workflow as workflowtype } from "~/api/instances/schema";

const exampleWorkflowInitial = {
  states: {
    stateFirst: {
      id: "hello-world",
      type: "function",
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

    stateSecond: {
      id: "exit",
      type: "function",
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
  },

  start: { state: "hello-world" },

  functions: [],
};

const exampleWorkflowStepOne = {
  states: {
    stateFirst: {
      id: "hello-world",
      type: "function",
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

    stateSecond: {
      id: "exit",
      type: "function",
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
  },

  start: { state: "hello-world" },

  functions: [],
};

const exampleWorkflowStepTwo = {
  states: {
    stateFirst: {
      id: "hello-world",
      type: "function",
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

    stateSecond: {
      id: "exit",
      type: "function",
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
  },

  start: { state: "hello-world" },

  functions: [],
};

const exampleWorkflowPending = {
  states: {
    stateFirst: {
      id: "hello-world",
      type: "function",
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

    stateSecond: {
      id: "exit",
      type: "function",
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
  },

  start: { state: "hello-world" },

  functions: [],
};

const exampleWorkflowFailed = {
  states: {
    stateFirst: {
      id: "stateFirst",
      type: "function",
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

    stateSecond: {
      id: "stateSecond",
      type: "function",
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

    stateThird: {
      id: "stateThird",
      type: "function",
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
  },

  start: { state: "stateFirst" },

  functions: [],
};

const exampleWorkflowComplete = {
  states: {
    stateFirst: {
      id: "hello-world",
      type: "function",
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

    stateSecond: {
      id: "exit",
      type: "function",
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
  },

  start: { state: "hello-world" },

  functions: [],
};

const exampleWorkflowComplex = {
  states: {
    stateFirst: {
      id: "hello-world",
      type: "function",
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
    stateSecondOptionA: {
      id: "actionOne",
      type: "function",
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
    stateSecondOptionB: {
      id: "actionTwo",
      type: "function",
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

    stateSecond: {
      id: "exit",
      type: "function",
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
  },

  start: { state: "hello-world" },

  functions: [],
};

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
    flow: {
      description:
        "Array of executed / visited states in an instance. Example - ['stateA', 'stateB']",
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
    <WorkflowDiagram workflow={exampleWorkflowInitial} flow={[]} />
  </div>
);

export const WorkflowInstancePending = () => (
  <div className="h-96">
    <WorkflowDiagram
      workflow={exampleWorkflowPending}
      flow={["stateFirst", "stateFirst"]}
      instanceStatus="pending"
    />
  </div>
);

export const WorkflowInstanceFailed = () => (
  <div className="h-96">
    <WorkflowDiagram
      workflow={exampleWorkflowFailed}
      flow={["stateFirst", "stateSecond", "stateThird"]}
      instanceStatus="failed"
    />
  </div>
);

export const WorkflowInstanceComplete = () => (
  <div className="h-96">
    <WorkflowDiagram
      workflow={exampleWorkflowComplete}
      flow={["helloworld", "exit"]}
      instanceStatus="complete"
    />
  </div>
);

export const UpdateWorkflow = () => {
  const [workflow, setWorkflow] = useState<workflowtype>(
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
        workflow={workflow}
        flow={["stateFirst"]}
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
    <WorkflowDiagram
      workflow={exampleWorkflowComplex}
      flow={["state"]}
      instanceStatus="complete"
    />
  </div>
);
