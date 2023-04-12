import { Bug, Eye, Filter } from "lucide-react";
import type { Meta, StoryObj } from "@storybook/react";
import Button from "../Button";
import { Toggle } from "./index";

const meta = {
  title: "Components/Toggle",
  component: Toggle,
} satisfies Meta<typeof Toggle>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <Toggle aria-label="Toggle italic" {...args}>
      <Filter className="h-4 w-4" />
    </Toggle>
  ),
  argTypes: {
    outline: {
      description: "with outline",
      type: {
        name: "boolean",
        required: false,
      },
    },
    size: {
      description: "size",
      control: "select",
      options: ["sm", "lg"],
      type: { name: "string", required: false },
    },
  },
};

export const OutlineToggle = () => (
  <Toggle aria-label="Toggle italic" outline>
    <Filter className="h-4 w-4" />
  </Toggle>
);

export const ToggleSize = () => (
  <div className="flex h-5 items-center space-x-4">
    <Toggle aria-label="Toggle italic" size="sm">
      <Filter className="h-4 w-4" />
    </Toggle>
    <Toggle aria-label="Toggle italic">
      <Filter className="h-4 w-4" />
    </Toggle>
    <Toggle aria-label="Toggle italic" size="lg">
      <Filter className="h-4 w-4" />
    </Toggle>
  </div>
);

export const DefaultPressed = () => (
  <Toggle aria-label="Toggle italic" defaultPressed>
    <Filter className="h-4 w-4" />
  </Toggle>
);

export const Toolbar = () => (
  <div className="flex flex-col space-y-3">
    <div className="flex space-x-1">
      <Toggle aria-label="Toggle italic" size="sm" outline>
        <Filter className="h-4 w-4" />
      </Toggle>
      <Toggle aria-label="Toggle italic" size="sm" outline>
        <Bug className="h-4 w-4" />
      </Toggle>
      <Toggle aria-label="Toggle italic" size="sm" outline>
        <Eye className="h-4 w-4" />
      </Toggle>
      <Button variant="outline" size="sm">
        Small Large Toolbar
      </Button>
    </div>
    <div className="flex space-x-1">
      <Toggle aria-label="Toggle italic" outline>
        <Filter className="h-4 w-4" />
      </Toggle>
      <Toggle aria-label="Toggle italic" outline>
        <Bug className="h-4 w-4" />
      </Toggle>
      <Toggle aria-label="Toggle italic" outline>
        <Eye className="h-4 w-4" />
      </Toggle>
      <Button variant="outline">Default Toolbar</Button>
    </div>
    <div className="flex space-x-1">
      <Toggle aria-label="Toggle italic" size="lg" outline>
        <Filter className="h-4 w-4" />
      </Toggle>
      <Toggle aria-label="Toggle italic" size="lg" outline>
        <Bug className="h-4 w-4" />
      </Toggle>
      <Toggle aria-label="Toggle italic" size="lg" outline>
        <Eye className="h-4 w-4" />
      </Toggle>
      <Button variant="outline" size="lg">
        Large Toolbar
      </Button>
    </div>
  </div>
);
