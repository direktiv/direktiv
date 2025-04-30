import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../../Dialog";

import Button from "~/design/Button";
import { FC } from "react";
import Input from "~/design/Input";

export const EditModal: FC = () => (
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Edit Text</DialogTitle>
    </DialogHeader>

    <fieldset className="gap-5">
      <label className="w-[150px] text-right" htmlFor="name">
        Content
      </label>
      <Input
        id="name"
        data-testid="variable-name"
        placeholder="Please insert the text"
      />
    </fieldset>

    <DialogFooter>
      <DialogClose asChild>
        <Button variant="outline" type="submit">
          Close
        </Button>
      </DialogClose>
      <Button variant="primary" type="submit">
        Save
      </Button>
    </DialogFooter>
  </DialogContent>
);
