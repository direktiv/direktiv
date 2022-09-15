import { ComponentMeta, ComponentStory } from '@storybook/react';
import '../../App.css';

import HelpIcon from './index';

export default {
  title: 'Components/HelpIcon',
  component: HelpIcon,
} as ComponentMeta<typeof HelpIcon>;

const Template: ComponentStory<typeof HelpIcon> = (args) => {
  return (<HelpIcon {...args} />)
};

export const NoHelpText = Template.bind({});

export const UpdateWorkflowTooltip = Template.bind({});
UpdateWorkflowTooltip.args = {
    msg: "Update a Workflow."
};