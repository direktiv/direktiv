import Button from "~/design/Button";
import { ComponentPropsWithoutRef } from "react";
import { DialogTrigger } from "~/design/Dialog";
import { PlusCircle } from "lucide-react";

const CreateItemButton = ({
  children,
  ...props
}: ComponentPropsWithoutRef<"button">) => (
  <DialogTrigger {...props} asChild>
    <Button variant="outline">
      <PlusCircle />
      {children}
    </Button>
  </DialogTrigger>
);

export default CreateItemButton;
