import { ComponentMeta, ComponentStory } from '@storybook/react';
import '../../App.css';
import './style.css'
import FlexBox from '../flexbox';

import Logs, { LogItem } from './logs';

const exampleInstanceLogs: LogItem[] = [
  {
    "t": "2022-09-08T03:14:20.937850Z",
    "msg": "Preparing workflow triggered by api."
  },
  {
    "t": "2022-09-08T03:14:20.943963Z",
    "msg": "Running state logic (step:1) -- helloworld"
  },
  {
    "t": "2022-09-08T03:14:20.944829Z",
    "msg": "\"Very long line. File = VGhpcyBtZXNzYWdlIGlzIGEgbGluayB0byB5b3V0dWJlLCBhbmQgaXMganVzdCBoZXJlIHRvIG1ha2UgdGhlIGxvZyBsaW5lIGxvbmdlci4gSWdub3JlIHRoZSBsaW5rOiBodHRwczovL3d3dy55b3V0dWJlLmNvbS93YXRjaD92PWRRdzR3OVdnWGNR\""
  },
  {
    "t": "2022-09-08T03:14:20.945668Z",
    "msg": "Transforming state data."
  },
  {
    "t": "2022-09-08T03:14:20.948874Z",
    "msg": "Workflow completed."
  }
]

export default {
  title: 'Components/Logs',
  component: Logs,
} as ComponentMeta<typeof Logs>;

const Template: ComponentStory<typeof Logs> = (args) => {
  return (<FlexBox style={{height: "250px"}}>
    <Logs {...args} />
  </FlexBox>)
};

export const LoadingDataCustomMessage = Template.bind({});
LoadingDataCustomMessage.args = {
  overrideLoadingMsg: "Loading Instance Logs"
};

export const NoDataCustomMessage = Template.bind({});
NoDataCustomMessage.args = {
  logItems: [],
  overrideNoDataMsg: "No Instance Logs"
};
NoDataCustomMessage.story = {
  parameters: {
      docs: {
          description: {
            story: 'Display custom message when logItems has no elements',
          },
        },
  }
};

export const InstanceLogs = Template.bind({});
InstanceLogs.args = {
  logItems: exampleInstanceLogs,
};
InstanceLogs.story = {
  parameters: {
      docs: {
          description: {
            story: 'Display example instance logs.',
          },
        },
  }
};

export const LogsWordWrap = Template.bind({});
LogsWordWrap.args = {
  logItems: exampleInstanceLogs,
  wordWrap: true
};
LogsWordWrap.story = {
  parameters: {
      docs: {
          description: {
            story: 'Display example instance logs and word wrap long lines.',
          },
        },
  }
};