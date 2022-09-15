import { ComponentMeta, ComponentStory } from '@storybook/react';
import '../../App.css';
import FlexBox from '../flexbox';

import WorkflowDiagram from './index';

const exampleWorkflow = `description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transition: exit
  transform:
    result: Hello world!
- id: exit
  type: noop
`

export default {
    title: 'Components/WorkflowDiagram',
    component: WorkflowDiagram,
    argTypes: {
        instanceStatus: {
            options: ['complete', 'failed', 'pending'],
            control: { type: 'select' },
            defaultValue: 'pending',
            description: "Status of current instance. This is used to display if flow is complete with animated connections.",
            table: {
                type: { summary: 'string' },
            }
        },
        workflow: {
            description: "YAML string of workflow.",
            table: {
                type: { summary: 'string' },
            }
        },
        flow: {
            description: "Array of executed states in an instance. Example - ['noopA', 'noopB']",
            table: {
                type: { summary: 'string[]' },
            }
        },
        disabled: {
            description: "Disables diagram zoom-in",
            table: {
                type: { summary: 'boolean' },
            }
        }
    }
};

const Template = (args) => {
    return (
        <FlexBox style={{ height: "380px" }}>
            <WorkflowDiagram {...args} />
        </FlexBox>
    )
};

export const Workflow = Template.bind({});
Workflow.args = {
    workflow: exampleWorkflow,
};

export const WorkflowInstancePending = Template.bind({});
WorkflowInstancePending.args = {
    workflow: exampleWorkflow,
    flow: ["helloworld"],
    instanceStatus: "pending"
};

WorkflowInstancePending.story = {
    parameters: {
        docs: {
            description: {
              story: 'Example of diagram when used in the context of an executing instance.',
            },
          },
    }
};

export const WorkflowInstanceComplete = Template.bind({});
WorkflowInstanceComplete.args = {
    workflow: exampleWorkflow,
    flow: ["helloworld", "exit"],
    instanceStatus: "complete"
};

WorkflowInstanceComplete.story = {
    parameters: {
        docs: {
            description: {
              story: 'Example of diagram when used in the context of a completed instance.',
            },
          },
    }
};