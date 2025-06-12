import { Dialog, DialogContent } from "~/design/Dialog";

import { BlockDeleteForm } from "./Delete";
import { BlockForm } from "..";
import { getPlaceholderBlock } from "../../PageCompiler/context/utils";
import { useBlockDialog } from "../BlockDialogProvider";
import { usePageEditor } from "../../PageCompiler/context/pageCompilerContext";

export const BlockDialog = () => {
  const { dialog, setDialog } = useBlockDialog();
  const { deleteBlock } = usePageEditor();

  if (!dialog) {
    return null;
  }

  /**
   * This handler is only used for closing the dialog. For opening a dialog,
   * we add custom onClick events to the trigger buttons.
   */
  const handleOnOpenChange = (open: boolean) => {
    if (open === false) {
      setDialog(null);
    }
  };

  const { action, block, blockType, path } = dialog;

  return (
    <Dialog open={!!dialog} onOpenChange={handleOnOpenChange}>
      <DialogContent className="z-50">
        {dialog.action === "edit" && (
          <BlockForm block={block} action={dialog.action} path={path} />
        )}
        {dialog.action === "create" && (
          <BlockForm
            block={getPlaceholderBlock(blockType)}
            action={action}
            path={path}
          />
        )}
        {dialog.action === "delete" && (
          <BlockDeleteForm
            type={block.type}
            action={dialog.action}
            path={path}
            onSubmit={(path) => deleteBlock(path)}
          />
        )}
      </DialogContent>
    </Dialog>
  );
};
