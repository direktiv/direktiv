import Button from "~/design/Button";
import { DialogTrigger } from "~/design/Dialog";
import { PlusCircle } from "lucide-react";

interface CreateItemButtonProps {
  onClick: () => void;
}

const CreateItemButton = (props: CreateItemButtonProps) => (
  <DialogTrigger {...props} asChild>
    <Button variant="outline">
      <PlusCircle />
    </Button>
  </DialogTrigger>
);

export default CreateItemButton;
