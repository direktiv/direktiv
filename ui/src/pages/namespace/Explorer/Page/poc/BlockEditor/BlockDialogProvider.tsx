import { createContext, useContext, useState } from "react";

import { AllBlocksType } from "../schema/blocks";
import { BlockDialogContent } from "./components/BlockDialogContent";
import { BlockPathType } from "../PageCompiler/Block";
import { Dialog } from "~/design/Dialog";

type DialogState = null | {
  action: "create" | "edit" | "delete";
  blockType: AllBlocksType["type"];
  block: AllBlocksType;
  path: BlockPathType;
};

type DialogContextType =
  | {
      dialog: DialogState;
      setDialog: React.Dispatch<React.SetStateAction<DialogState>>;
    }
  | undefined;

const DialogContext = createContext<DialogContextType>(undefined);

export const BlockDialogProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const [dialog, setDialog] = useState<DialogState>(null);
  /**
   * This handler is only used for closing the dialog. For opening a dialog,
   * we add custom onClick events to the trigger buttons.
   */
  const handleOnOpenChange = (open: boolean) => {
    if (open === false) {
      setDialog(null);
    }
  };
  return (
    <DialogContext.Provider value={{ dialog, setDialog }}>
      <Dialog open={!!dialog} onOpenChange={handleOnOpenChange}>
        {children}
        <BlockDialogContent />
      </Dialog>
    </DialogContext.Provider>
  );
};

export const useBlockDialog = () => {
  const context = useContext(DialogContext);
  if (!context)
    throw new Error("useBlockDialog must be used within BlockDialogProvider");
  return context;
};
