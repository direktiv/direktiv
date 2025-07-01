import { BlockDeleteForm } from "./Delete";
import { BlockForm } from "..";
import { usePageEditor } from "../../PageCompiler/context/pageCompilerContext";
import { usePageEditorPanel } from "../EditorPanelProvider";

export const EditorPanel = () => {
  const { deleteBlock } = usePageEditor();
  const { panel, setPanel } = usePageEditorPanel();

  if (!panel) {
    return (
      <div>
        Placeholder for a global page form, including page settings and drag and
        drop sources for adding blocks.
      </div>
    );
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
