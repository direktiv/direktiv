import { BlockDeleteForm } from "./Delete";
import { BlockForm } from "..";
import { usePageEditor } from "../../PageCompiler/context/pageCompilerContext";
import { usePageEditorPanel } from "../EditorPanelProvider";

export const EditorPanel = () => {
  const { deleteBlock } = usePageEditor();
  const { panel, setPanel } = usePageEditorPanel();

  if (!panel) {
    // Instead of nothing, we could later display global page settings.
    return null;
  }

  return (
    <>
      {panel.action === "delete" ? (
        <BlockDeleteForm
          path={panel.path}
          onSubmit={() => {
            setPanel(null);
            deleteBlock(panel.path);
          }}
        />
      ) : (
        <BlockForm
          action={panel.action}
          path={panel.path}
          block={panel.block}
        />
      )}
    </>
  );
};
