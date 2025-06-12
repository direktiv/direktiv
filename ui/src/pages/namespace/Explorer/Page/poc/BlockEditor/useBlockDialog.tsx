import { AllBlocksType } from "../schema/blocks";
import { useState } from "react";

type DialogState = null | {
  action: "create" | "edit" | "delete";
  blockType: AllBlocksType["type"];
};

export const useBlockDialog = () => {
  const [dialog, setDialog] = useState<DialogState>(null);
  return { dialog, setDialog };
};
