import { BlockForm } from "..";
import { usePageEditorPanel } from "../EditorPanelProvider";

export const EditorPanel = () => {
  const { panel } = usePageEditorPanel();

  if (!panel) {
    return (
      <div>
        Placeholder for a global page form, including page settings and drag and
        drop sources for adding blocks.
      </div>
    );
  }

  return (
    <BlockForm action={panel.action} path={panel.path} block={panel.block} />
  );
};
