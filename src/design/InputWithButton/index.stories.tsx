
import type { Meta, StoryObj } from "@storybook/react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../Tooltip";
import Button from "../Button";
import { InputWithButton } from "./index";
import { Toggle } from "../Toggle";
import Input from "../Input";
import { ZoomIn } from "lucide-react";

const meta = {
  title: "Components/InputWithButton",
  component: InputWithButton,
} satisfies Meta<typeof InputWithButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <InputWithButton {...args}>
      <Input />
      <Button icon >
        <ZoomIn />
      </Button>
    </InputWithButton>
  ),
};
