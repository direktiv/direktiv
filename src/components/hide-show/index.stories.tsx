import { ComponentMeta, ComponentStory } from '@storybook/react';
import '../../App.css';
import './style.css';

import HideShowButton from './index';

export default {
  title: 'Components/HideShowButton',
  component: HideShowButton,
} as ComponentMeta<typeof HideShowButton>;

const Template: ComponentStory<typeof HideShowButton> = (args) => {
  return (<HideShowButton {...args}/>)
};

export const Password = Template.bind({});
Password.args = {
  show: false,
  field: "password",
};