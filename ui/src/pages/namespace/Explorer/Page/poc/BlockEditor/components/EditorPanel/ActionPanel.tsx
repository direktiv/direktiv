import { BlockForm } from "../..";
import { EditorPanelAction } from "../../EditorPanelProvider";

export const ActionPanel = ({ panel }: { panel: EditorPanelAction }) => (
  <div className="h-[300px] overflow-y-scroll border-b-2 border-gray-4 p-3 dark:border-gray-dark-4 sm:h-[calc(100vh-230px)] sm:border-b-0 sm:border-r-2">
    <BlockForm action={panel.action} path={panel.path} block={panel.block} />
  </div>
);
