import type { Meta, StoryObj } from "@storybook/react";
import { Switch } from "./index";

const meta = {
  title: "Components/Switch",
  component: Switch,
} satisfies Meta<typeof Switch>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => <Switch {...args} />,
  argTypes: {
    disabled: {
      description: "Disable/Enable",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
  },
};

export const DisabledSwitch = () => <Switch disabled />;
