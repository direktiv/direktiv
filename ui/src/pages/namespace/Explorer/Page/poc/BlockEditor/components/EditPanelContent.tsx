import { BlockDeleteForm } from "./Delete";
import { BlockForm } from "..";
import { DialogContent } from "~/design/Dialog";
import { useBlockDialog } from "../EditPanelProvider";
import { usePageEditor } from "../../PageCompiler/context/pageCompilerContext";

export const EditPanel = () => {
  const { panel } = useBlockDialog();
  const { deleteBlock } = usePageEditor();

  if (!panel) {
    return null;
  }

  const { action, block, path } = panel;

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
