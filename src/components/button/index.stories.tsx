import { ComponentMeta, ComponentStory } from '@storybook/react';
import '../../App.css';

import Button from './index';

export default {
  title: 'Components/Button',
  component: Button,
} as ComponentMeta<typeof Button>;

const Template: ComponentStory<typeof Button> = (args) => {
  return (<Button {...args} />)
};

export const DefaultButton = Template.bind({});
DefaultButton.args = {
  children: "Button",
  tooltip: "Tooltip",
};

export const DisabledButton = Template.bind({});
DisabledButton.args = {
  children: "Disabled Button",
  disabledTooltip: "Seperate tooltip for disabled state.",
  disabled: true
};
DisabledButton.story = {
  parameters: {
      docs: {
          description: {
            story: 'Disabled Button example with disabled tooltip prop set.',
          },
        },
  }
};

export const InfoButton = Template.bind({});
InfoButton.args = {
  children: "Info Button",
  tooltip: "Button variant outlined and color info",
  variant: "outlined",
  color: "info"
};
InfoButton.story = {
  parameters: {
      docs: {
          description: {
            story: 'Info Button example.',
          },
        },
  }
};

export const SynchronousButton = Template.bind({});
SynchronousButton.args = {
  children: "Synchronous Button",
  tooltip: "Once clicked, will be set to disabled until onClick function has finished.",
  disabledTooltip: "Waiting 2 seconds",
  onClick: async () => {
    // Wait for 2 seconds
    return new Promise(resolve => setTimeout(resolve, 2000))
  },
  asyncDisable: true
};
SynchronousButton.story = {
  parameters: {
      docs: {
          description: {
            story: 'When setting asyncDisable to true, buttons will be set to disabled when the onClick event is fired. Disabled will automatically be set back to false when all promises in the onClick callback are resolved. In this example a 2000ms timeout promise is executed on the onClick event.',
          },
        },
  }
};