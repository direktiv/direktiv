import type { Meta, StoryObj } from "@storybook/react";
import Alert from "./index";

const meta = {
  title: "Components/Alert",
  component: Alert,
} satisfies Meta<typeof Alert>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => <Alert {...args}>Alert</Alert>,
  argTypes: {
    variant: {
      description: "Variant of the alert",
      options: ["info", "success", "warning", "error"],
      control: { type: "select" },
      type: "string",
    },
  },
};

export const AlertVariants = () => (
  <div className="flex flex-col space-y-5">
    <Alert>
      default alert text here default alert text here default alert text here
      default alert text here default alert text here default alert text here
    </Alert>
    <Alert variant="info">info alert text here</Alert>
    <Alert variant="success">success alert text here </Alert>
    <Alert variant="warning">warning alert text here </Alert>
    <Alert variant="error">error alert text here </Alert>
  </div>
);
