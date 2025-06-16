import { BlockDeleteForm } from "./Delete";
import { BlockForm } from "..";
import { DialogContent } from "~/design/Dialog";
import { useBlockDialog } from "../BlockDialogProvider";
import { usePageEditor } from "../../PageCompiler/context/pageCompilerContext";

export const BlockDialogContent = () => {
  const { dialog } = useBlockDialog();
  const { deleteBlock } = usePageEditor();

  if (!dialog) {
    return null;
  }

  const { action, block, path } = dialog;

  return (
    <DialogContent className="z-50">
      {action === "delete" ? (
        <BlockDeleteForm
          type={block.type}
          action={action}
          path={path}
          onSubmit={(path) => deleteBlock(path)}
        />
      ) : (
        <BlockForm block={block} action={action} path={path} />
      )}
    </DialogContent>
  );
};
