import { createContext, useContext, useState } from "react";

import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { EditPanel } from "./components/EditPanelContent";

type EditorPanelState = null | {
  action: "create" | "edit" | "delete";
  block: AllBlocksType;
  path: BlockPathType;
};

type EditorPanelContextType =
  | {
      panel: EditorPanelState;
      setPanel: React.Dispatch<React.SetStateAction<EditorPanelState>>;
    }
  | undefined;

const EditorPanelContext = createContext<EditorPanelContextType>(undefined);

export const EditorPanelLayoutProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const [panel, setPanel] = useState<EditorPanelState>(null);
  /**
   * This handler is only used for closing the dialog. For opening a dialog,
   * we add custom onClick events to the trigger buttons.
   */

  return (
    <EditorPanelContext.Provider value={{ panel, setPanel }}>
      <div className="flex">
        <div className="w-1/3 max-w-xs">
          <EditPanel />
        </div>
        <div className="w-full">{children}</div>
      </div>
    </EditorPanelContext.Provider>
  );
};

export const useEditorPanel = () => {
  const context = useContext(EditorPanelContext);
  if (!context)
    throw new Error(
      "useEditorPanel must be used within EditorPanelLayoutProvider"
    );
  return context;
};
