import { createContext, useContext, useState } from "react";

import { AllBlocksType } from "../schema/blocks";
import { BlockPathType } from "../PageCompiler/Block";

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
  return (
    <DialogContext.Provider value={{ dialog, setDialog }}>
      {children}
    </DialogContext.Provider>
  );
};

export const useBlockDialog = () => {
  const context = useContext(DialogContext);
  if (!context)
    throw new Error("useBlockDialog must be used within BlockDialogProvider");
  return context;
};
