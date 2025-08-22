import { BlockForm } from "../..";
import { EditorPanelAction } from "../../EditorPanelProvider";
import { PanelContainer } from "./PanelContainer";

export const ActionPanel = ({ panel }: { panel: EditorPanelAction }) => (
  <PanelContainer className="overflow-y-scroll p-3">
    <BlockForm action={panel.action} path={panel.path} block={panel.block} />
  </PanelContainer>
);
