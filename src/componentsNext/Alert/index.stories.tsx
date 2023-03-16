import type { Meta, StoryObj } from "@storybook/react";
import Alert from "./index";

const meta = {
  title: "Components (next)/Alert",
  component: Alert,
} satisfies Meta<typeof Alert>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => {
    return <Alert {...args} />;
  },
  argTypes: {
    variant: {
      description: "Variant of the alert",
      options: ["info", "success", "warning", "error", "default"],
      control: { type: "select" },
      type: "string",
    },
    text: {
      description: "Text content of the alert",
      control: {
        type: "text",
        defaultValue: "Hey, this is alert",
      },
      type: { name: "string", required: false },
    },
  },
};

export const AlertVariants = () => (
  <div className="flex space-x-5">
    <Alert variant="info" text="info alert text here" />
    <Alert variant="success" text="success alert text here" />
    <Alert variant="warning" text="warning alert text here" />
    <Alert variant="error" text="error alert text here" />
    <Alert variant="default" text="default alert text here" />
  </div>
)