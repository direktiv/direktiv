import { DialogHeader, DialogTitle } from "~/design/Dialog";

import { BlockEditFormProps } from ".";

export const Text = ({ block, path }: BlockEditFormProps) => (
  <>
    <DialogHeader>
      <DialogTitle>Edit Text Block @ {path.join(".")}</DialogTitle>
    </DialogHeader>
    <div>{JSON.stringify(block)}</div>
  </>
);
