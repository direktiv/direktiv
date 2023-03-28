import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "./index";
import type { Meta, StoryObj } from "@storybook/react";
import Button from "../Button";

const meta = {
  title: "Components/Dialog",
  component: Dialog,
} satisfies Meta<typeof Dialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <>
      <Dialog>
        <DialogTrigger>Open</DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Dialog Title</DialogTitle>
            <DialogDescription>
              This is a description of the dialog.
            </DialogDescription>
          </DialogHeader>
        </DialogContent>
      </Dialog>
    </>
  ),
  tags: ["autodocs"],
  argTypes: {},
};

export const WithButtonAsTrigger = () => (
  <Dialog>
    <DialogTrigger asChild>
      <Button>Use asChild to use your own button</Button>
    </DialogTrigger>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Dialog Title</DialogTitle>
        <DialogDescription>
          This is a description of the dialog.
        </DialogDescription>
      </DialogHeader>
    </DialogContent>
  </Dialog>
);
