import { createContext, useContext, useState } from "react";

import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { EditorPanel } from "./components/EditorPanelContent";
import { usePageStateContext } from "../PageCompiler/context/pageCompilerContext";

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
  const { mode } = usePageStateContext();

  if (mode === "edit") {
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
  }

  return <>{children}</>;
};

export const usePageEditorPanel = () => {
  const context = useContext(EditorPanelContext);
  if (!context)
    throw new Error(
      "usePageEditorPanel must be used within EditorPanelLayoutProvider"
    );
  return context;
};
