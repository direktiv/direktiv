import { BlockForm } from "../..";
import { EditorPanelAction } from "../../EditorPanelProvider";
import { PanelContainer } from "./PanelContainer";

export const ActionPanel = ({ panel }: { panel: EditorPanelAction }) => (
  <PanelContainer className="h-[555px] overflow-y-auto">
    <BlockForm action={panel.action} path={panel.path} block={panel.block} />
  </PanelContainer>
);
