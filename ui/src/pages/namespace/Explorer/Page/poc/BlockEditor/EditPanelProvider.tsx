import { createContext, useContext, useState } from "react";

import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";
import { Dialog } from "~/design/Dialog";
import { EditPanel } from "./components/EditPanelContent";

type EditPanelState = null | {
  action: "create" | "edit" | "delete";
  block: AllBlocksType;
  path: BlockPathType;
};

type EditPanelContextType =
  | {
      panel: EditPanelState;
      setPanel: React.Dispatch<React.SetStateAction<EditPanelState>>;
    }
  | undefined;

const EditPanelContext = createContext<EditPanelContextType>(undefined);

export const EditPanelLayoutProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const [panel, setPanel] = useState<EditPanelState>(null);
  /**
   * This handler is only used for closing the dialog. For opening a dialog,
   * we add custom onClick events to the trigger buttons.
   */
  const handleOnOpenChange = (open: boolean) => {
    if (open === false) {
      setPanel(null);
    }
  };
  return (
    // Or I could do both panes here? And "Children" will only be the renderer.
    <EditPanelContext.Provider value={{ panel, setPanel }}>
      <Dialog open={!!panel} onOpenChange={handleOnOpenChange}>
        {children}
        {/* Block dialog content will move to another part of the layout */}
        <EditPanel />
      </Dialog>
    </EditPanelContext.Provider>
  );
};

export const useBlockDialog = () => {
  const context = useContext(EditPanelContext);
  if (!context)
    throw new Error("useBlockDialog must be used within BlockDialogProvider");
  return context;
};
