import { BlockDeleteForm } from "./Delete";
import { BlockForm } from "..";
import { DialogContent } from "~/design/Dialog";
import { getPlaceholderBlock } from "../../PageCompiler/context/utils";
import { useBlockDialog } from "../BlockDialogProvider";
import { usePageEditor } from "../../PageCompiler/context/pageCompilerContext";

export const BlockDialogContent = () => {
  const { dialog } = useBlockDialog();
  const { deleteBlock } = usePageEditor();

  if (!dialog) {
    return null;
  }

  const { action, block, blockType, path } = dialog;

  return (
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
  );
};
