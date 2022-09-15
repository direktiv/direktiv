import { ComponentMeta, ComponentStory } from '@storybook/react';
import '../../App.css';
import "./style.css"


import FlexBox from './index';

const itemStyle = { backgroundColor: "#2F8F9D", borderRadius: "6px", padding: "10px", textAlign: "center" } as React.CSSProperties

export default {
  title: 'Components/FlexBox',
  component: FlexBox,
  argTypes: {
    center: {
      options: ["y","x","xy", false],
      control: { type: 'select' },
      defaultValue: false,
    },
    gap: {
      options: ["md","sm", false],
      control: { type: 'select' },
      defaultValue: false,
    },
    hide: {
      control: { type: "boolean" },
    },
    col: {
      control: { type: "boolean" },
    },
    row: {
      control: { type: "boolean" },
    },
    tall: {
      control: { type: "boolean" },
    },
    wrap: {
      control: { type: "boolean" },
    },
  },
} as ComponentMeta<typeof FlexBox>;

const Template: ComponentStory<typeof FlexBox> = (args) => {
  return (
    <FlexBox {...args} >
      <div style={itemStyle}> Item 1 </div>
      <div style={{ ...itemStyle, backgroundColor: "#3BACB6" }}> Item 2 </div>
      <div style={{ ...itemStyle, backgroundColor: "#82DBD8" }}> Item 3 </div>
      <div style={{ ...itemStyle, backgroundColor: "#B3E8E5" }}> Item 4 </div>
    </FlexBox>
  )
};

export const Column = Template.bind({});
Column.args = {
  col: true,
  gap: true
};

export const Row = Template.bind({});
Row.args = {
  row: true,
  gap: true
};

export const Center = Template.bind({});
Center.args = {
  col: true,
  center: true,
  gap: true
};

export const NoGap = Template.bind({});
NoGap.args = {
  col: true,
};