import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "./index";
import type { Meta, StoryObj } from "@storybook/react";
import Button from "../Button";
import { DialogClose } from "@radix-ui/react-dialog";
import { Folder, Settings } from "lucide-react";

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
            <DialogTitle>
              <Settings />
              Dialog Title
            </DialogTitle>
            <DialogDescription>
              This is a description of the dialog.
            </DialogDescription>
          </DialogHeader>
          Content goes here
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="ghost">Cancel</Button>
            </DialogClose>
            <Button>Submit</Button>
          </DialogFooter>
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
        <DialogTitle>
          <Settings /> Dialog Title
        </DialogTitle>
        <DialogDescription>
          This is a description of the dialog.
        </DialogDescription>
      </DialogHeader>
      Content goes here
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Cancel</Button>
        </DialogClose>
        <Button>Submit</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
);
