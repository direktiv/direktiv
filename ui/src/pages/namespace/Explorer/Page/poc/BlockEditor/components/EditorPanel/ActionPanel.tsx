import { BlockForm } from "../..";
import { EditorPanelAction } from "../../EditorPanelProvider";

export const ActionPanel = ({ panel }: { panel: EditorPanelAction }) => (
  <BlockForm
    data-testid="editor-sidePanel"
    action={panel.action}
    path={panel.path}
    block={panel.block}
  />
);
