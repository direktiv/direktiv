import Button from "~/design/Button";
import { ComponentPropsWithoutRef } from "react";
import { DialogTrigger } from "~/design/Dialog";
import { PlusCircle } from "lucide-react";

const CreateItemButton = (props: ComponentPropsWithoutRef<"button">) => (
  <DialogTrigger {...props} asChild>
    <Button variant="outline">
      <PlusCircle />
    </Button>
  </DialogTrigger>
);

export default CreateItemButton;
