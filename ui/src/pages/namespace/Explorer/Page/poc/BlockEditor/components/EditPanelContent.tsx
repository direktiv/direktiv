import { BlockDeleteForm } from "./Delete";
import { BlockForm } from "..";
import { useEditorPanel } from "../EditorPanelProvider";
import { usePageEditor } from "../../PageCompiler/context/pageCompilerContext";

export const EditPanel = () => {
  const { panel } = useEditorPanel();
  const { deleteBlock } = usePageEditor();

  if (!panel) {
    return null;
  }

  const { action, block, path } = panel;

  return (
    <>
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
    </>
  );
};
