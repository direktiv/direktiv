import { ComponentMeta, ComponentStory } from '@storybook/react';
import '../../App.css';

import InvalidWorkflow from './index';

const exampleInvalidWorkflowError = `can not read a block mapping entry; a multiline key may not be an implicit key (5:5)

2 | states:
3 | - id: helloworld
4 |   typ
5 |    e: noop
---------^
6 |   transform:
7 |     result: Hello world!`

export default {
  title: 'Components/InvalidWorkflow',
  component: InvalidWorkflow,
} as ComponentMeta<typeof InvalidWorkflow>;

const Template: ComponentStory<typeof InvalidWorkflow> = (args) => {
  return (<InvalidWorkflow {...args} />)
};

export const ErrorExample = Template.bind({});
ErrorExample.args = {
  invalidWorkflow: exampleInvalidWorkflowError,
};