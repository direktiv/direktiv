import { ComponentMeta, ComponentStory } from '@storybook/react';
import '../../App.css';

import Loader from './index';

export default {
  title: 'Components/Loader',
  component: Loader,
} as ComponentMeta<typeof Loader>;

const Template: ComponentStory<typeof Loader> = (args) => {
  return (<Loader {...args} />)
};

export const Loading = Template.bind({});
Loading.args = {
    load: true
};

export const LoadAfter5Sec = Template.bind({});
LoadAfter5Sec.args = {
    load: true,
    timer: 5000
};