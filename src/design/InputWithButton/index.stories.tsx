import type { Meta, StoryObj } from "@storybook/react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../Tooltip";
import Button from "../Button";
import CopyButton from "../CopyButton";
import { CopyIcon } from "lucide-react";
import Input from "../Input";
import { InputWithButton } from "./index";
import { useState } from "react";

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

export const IconWithToolTip = () => {
  const [value, setValue] = useState("some value");
  return (
    <TooltipProvider>
      <InputWithButton>
        <Input
          value={value}
          onChange={(e) => {
            setValue(e.target.value);
          }}
        />
        <Tooltip>
          <TooltipTrigger>
            <CopyButton
              value={value}
              buttonProps={{
                icon: true,
                variant: "ghost",
              }}
            />
          </TooltipTrigger>
          <TooltipContent>Copy Value</TooltipContent>
        </Tooltip>
      </InputWithButton>
    </TooltipProvider>
  );
};
