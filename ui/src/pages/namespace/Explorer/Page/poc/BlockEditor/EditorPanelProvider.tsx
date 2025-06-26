import { createContext, useContext, useState } from "react";

import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { EditorPanel } from "./components/EditorPanelContent";

type EditorPanelState =
  | null
  | {
      action: "delete";
      path: BlockPathType;
    }
  | { action: "create" | "edit"; block: AllBlocksType; path: BlockPathType };

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
      <div className="flex gap-5">
        <div className="w-1/3 max-w-md shrink-0 overflow-x-hidden">
          <EditorPanel />
        </div>
        <div className="min-w-0 flex-1">{children}</div>
      </div>
    </EditorPanelContext.Provider>
  );
};

export const usePageEditorPanel = () => {
  const context = useContext(EditorPanelContext);
  if (!context)
    throw new Error(
      "usePageEditorPanel must be used within EditorPanelLayoutProvider"
    );
  return context;
};
