import { BlockDeleteForm } from "./Delete";
import { BlockForm } from "..";
import { DialogContent } from "~/design/Dialog";
import { useBlockDialog } from "../BlockDialogProvider";

export const BlockDialogContent = () => {
  const { dialog } = useBlockDialog();

  return (
    <DialogContent className="z-50">
      {dialog?.action === "delete" ? <BlockDeleteForm /> : <BlockForm />}
    </DialogContent>
  );
};
