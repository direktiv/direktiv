import {
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogXClose,
} from "~/design/Dialog";

import { LocalDialog, LocalDialogContent } from ".";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { LocalDialogContainer } from "./container";
import type { Meta } from "@storybook/react";

const meta = {
  title: "Components/LocalDialog",
  tags: ["autodocs"],
  component: LocalDialog,
  argTypes: {},
} satisfies Meta<typeof LocalDialog>;

export default meta;

export const Default = () => (
  <div className="flex h-[50vh] w-full flex-row gap-5">
    <Card className="w-1/3 max-w-md shrink-0 overflow-x-hidden p-5">
      This Card is outside the container and not blocked by the dialog.
    </Card>
    <LocalDialogContainer className="min-w-0 flex-1">
      <Card className="p-5">
        <div className="mb-5">
          This Card is in the container and blocked by the dialog when opened.
        </div>
        <LocalDialog>
          <DialogTrigger asChild>
            <Button>Open Dialog</Button>
          </DialogTrigger>
          <LocalDialogContent>
            <DialogXClose />
            <DialogHeader>
              <DialogTitle>Dialog Title</DialogTitle>
            </DialogHeader>
            <div>Some dialog content</div>
          </LocalDialogContent>
        </LocalDialog>
      </Card>
    </LocalDialogContainer>
  </div>
);
