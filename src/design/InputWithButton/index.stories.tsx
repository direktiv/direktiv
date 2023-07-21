import type { Meta, StoryObj } from "@storybook/react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../Tooltip";
import Button from "../Button";
import { CopyIcon } from "lucide-react";
import Input from "../Input";
import { InputWithButton } from "./index";

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
      <Button icon variant="ghost">
        <CopyIcon />
      </Button>
    </InputWithButton>
  ),
};

export const InputWithTextButton = () => (
  <InputWithButton>
    <Input />
    <Button>Show Password</Button>
  </InputWithButton>
);

export const IconWithToolTip = () => (
  <TooltipProvider>
    <InputWithButton>
      <Input />
      <Tooltip>
        <TooltipTrigger>
          <Button icon variant="ghost">
            <CopyIcon />
          </Button>
        </TooltipTrigger>
        <TooltipContent>Copy Value</TooltipContent>
      </Tooltip>
    </InputWithButton>
  </TooltipProvider>
);
