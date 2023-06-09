import { ChevronDown, GitBranchIcon, Undo } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../Dropdown";
import type { Meta, StoryObj } from "@storybook/react";
import Button from "../Button";
import { ButtonBar } from "./index";

const meta = {
  title: "Components/ButtonBar",
  component: ButtonBar,
} satisfies Meta<typeof ButtonBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <ButtonBar {...args}>
      <Button>Start</Button>
      <Button>Mid-1</Button>
      <Button>Mid-2</Button>
      <Button>End</Button>
    </ButtonBar>
  ),
};

export const GitButtonBar = () => (
  <div className="flex gap-2">
    <DropdownMenu>
      <ButtonBar>
        <Button variant="outline">
          <GitBranchIcon /> Review
        </Button>
        <DropdownMenuTrigger asChild>
          <Button variant="outline">
            <ChevronDown />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem>
            <Undo className="mr-2 h-4 w-4" />
            Revert to Previous
          </DropdownMenuItem>
        </DropdownMenuContent>
      </ButtonBar>
    </DropdownMenu>
  </div>
);
