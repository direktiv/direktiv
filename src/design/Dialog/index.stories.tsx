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
import { Settings } from "lucide-react";
import { useState } from "react";

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
      <div className="my-3">
        Content goes here. A div with <strong>my-3</strong> is recommended.
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Cancel</Button>
        </DialogClose>
        <Button>Submit</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
);

export const ControlledDialogWithForm = () => {
  const [openDialog, setOpenDialog] = useState(false);
  const formId = "my-form";
  return (
    <Dialog open={openDialog} onOpenChange={setOpenDialog}>
      <DialogTrigger asChild>
        <Button>Controlled dialog with a form</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            <Settings /> Dialog Title
          </DialogTitle>
          <DialogDescription>
            In this example the dialogs open state is also controlled with the
            submit of the form. Please note that the submit can also be
            triggered with a button outside the form. This is helpful since the
            DialogContent should have DialogHeader and DialogFooter as direct
            children. You must use the form attribute on the button to submit
            and give it the id of the form you want to submit.
          </DialogDescription>
        </DialogHeader>
        <div className="my-3">
          <form id={formId} onSubmit={() => setOpenDialog(false)}>
            <fieldset className="flex items-center gap-5">
              <label className="w-[90px] text-right text-[14px]" htmlFor="name">
                Name
              </label>
              <input
                className="inline-flex h-[35px] w-full flex-1 items-center justify-center rounded-[4px] px-[10px] text-[14px] leading-none shadow-[0_0_0_1px] outline-none focus:shadow-[0_0_0_2px]"
                id="name"
                placeholder="just submit this"
              />
            </fieldset>
          </form>
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">Cancel</Button>
          </DialogClose>
          <Button type="submit" form={formId}>
            submit outside the form
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
