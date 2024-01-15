import { Bug, Eye, Filter } from "lucide-react";
import type { Meta, StoryObj } from "@storybook/react";
import Button from "../Button";
import { ButtonBar } from "../ButtonBar";
import { Toggle } from "./index";
import { useState } from "react";

const meta = {
  title: "Components/Toggle",
  component: Toggle,
} satisfies Meta<typeof Toggle>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <Toggle aria-label="Toggle italic" {...args}>
      <Filter />
    </Toggle>
  ),
  argTypes: {
    size: {
      description: "size",
      control: "select",
      options: ["sm", "lg"],
      type: { name: "string", required: false },
    },
  },
};

export const ToggleSize = () => (
  <div className="flex h-5 items-center space-x-4">
    <Toggle aria-label="Toggle italic" size="sm">
      <Filter />
    </Toggle>
    <Toggle aria-label="Toggle italic">
      <Filter />
    </Toggle>
    <Toggle aria-label="Toggle italic" size="lg">
      <Filter />
    </Toggle>
  </div>
);

export const WithTextAndState = () => {
  const [pressed, setPressed] = useState(false);
  return (
    <Toggle
      aria-label="Toggle italic"
      pressed={pressed}
      onPressedChange={setPressed}
    >
      <Filter /> {pressed ? "On" : "Off"}
    </Toggle>
  );
};

export const DefaultPressed = () => (
  <Toggle aria-label="Toggle italic" defaultPressed>
    <Filter />
  </Toggle>
);

export const Toolbar = () => (
  <div className="flex flex-col space-y-3">
    <ButtonBar>
      <Toggle aria-label="Toggle italic" size="sm">
        <Filter />
      </Toggle>
      <Toggle aria-label="Toggle italic" size="sm">
        <Bug />
      </Toggle>
      <Toggle aria-label="Toggle italic" size="sm">
        <Eye />
      </Toggle>
      <Button variant="outline" size="sm">
        <Eye /> Small Large Toolbar
      </Button>
    </ButtonBar>
    <ButtonBar>
      <Toggle aria-label="Toggle italic">
        <Filter />
      </Toggle>
      <Toggle aria-label="Toggle italic">
        <Bug />
      </Toggle>
      <Toggle aria-label="Toggle italic">
        <Eye />
      </Toggle>
      <Button variant="outline">
        <Eye /> Default Toolbar
      </Button>
    </ButtonBar>
    <ButtonBar>
      <Toggle aria-label="Toggle italic" size="lg">
        <Filter />
      </Toggle>
      <Toggle aria-label="Toggle italic" size="lg">
        <Bug />
      </Toggle>
      <Toggle aria-label="Toggle italic" size="lg">
        <Eye />
      </Toggle>
      <Button variant="outline" size="lg">
        <Eye />
        Large Toolbar
      </Button>
    </ButtonBar>
  </div>
);
