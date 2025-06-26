import { createContext, useContext, useState } from "react";

import { AllBlocksType } from "../schema/blocks";
import { BlockDialogContent } from "./components/BlockDialogContent";
import { BlockPathType } from "../PageCompiler/Block";
import { Dialog } from "~/design/Dialog";

type DialogState = null | {
  action: "create" | "edit" | "delete";
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

  return (
    <DialogContext.Provider value={{ dialog, setDialog }}>
      <Dialog open={!!dialog}>
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
