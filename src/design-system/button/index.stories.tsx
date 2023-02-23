import type { Meta, StoryObj } from "@storybook/react";

import Button from "./index";
import { VscAccount } from "react-icons/vsc";

const meta = {
  title: "Design System/Button",
  component: Button,
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => <Button {...args}>Button</Button>,
  args: {},
  argTypes: {
    size: {
      description: "Button size",
      control: "select",
      options: ["xs", "md", "lg"],
      type: { name: "string", required: false },
    },
    color: {
      description: "Button color",
      control: "select",
      options: [
        "primary",
        "secondary",
        "accent",
        "ghost",
        "link",
        "info",
        "success",
        "warning",
        "error",
      ],
      type: { name: "string", required: false },
    },
    outline: {
      description: "Outline button",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    active: {
      description: "Button in active state",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    loading: {
      description: "Button in loading state",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
  },
};

export const ButtonSizes = () => {
  return (
    <div className="flex flex-wrap gap-5">
      <Button size="xs">XS Button</Button>
      <Button size="sm">SM Button</Button>
      <Button>Normal Button</Button>
      <Button size="lg">lg Button</Button>
    </div>
  );
};

export const ButtonColors = () => (
  <div className="flex flex-wrap gap-5">
    <Button>Default</Button>
    <Button color="primary">Primary</Button>
    <Button color="secondary">Secondary</Button>
    <Button color="accent">Accent</Button>
    <Button color="ghost">Ghost</Button>
    <Button color="link">Link</Button>
    <Button color="info">Info</Button>
    <Button color="success">Success</Button>
    <Button color="warning">Warning</Button>
    <Button color="error">Error</Button>
  </div>
);

export const ActiveButtonColors = () => (
  <div className="flex flex-wrap gap-5">
    <Button active>Default</Button>
    <Button active color="primary">
      Primary
    </Button>
    <Button active color="secondary">
      Secondary
    </Button>
    <Button active color="accent">
      Accent
    </Button>
    <Button active color="ghost">
      Ghost
    </Button>
    <Button active color="link">
      Link
    </Button>
    <Button active color="info">
      Info
    </Button>
    <Button active color="success">
      Success
    </Button>
    <Button active color="warning">
      Warning
    </Button>
    <Button active color="error">
      Error
    </Button>
  </div>
);

export const Outline = () => (
  <div className="flex flex-wrap gap-5">
    <Button outline>Default</Button>
    <Button outline color="primary">
      Primary
    </Button>
    <Button outline color="secondary">
      Secondary
    </Button>
    <Button outline color="accent">
      Accent
    </Button>
    <Button outline color="ghost">
      Ghost
    </Button>
    <Button outline color="link">
      Link
    </Button>
    <Button outline color="info">
      Info
    </Button>
    <Button outline color="success">
      Success
    </Button>
    <Button outline color="warning">
      Warning
    </Button>
    <Button outline color="error">
      Error
    </Button>
  </div>
);

export const Loading = () => (
  <div className="flex flex-wrap gap-5">
    <Button outline loading>
      Loading
    </Button>
  </div>
);

export const WithIcon = () => (
  <div className="flex flex-wrap gap-5">
    <Button color="primary" className="gap-5">
      Loading <VscAccount />
    </Button>
  </div>
);
