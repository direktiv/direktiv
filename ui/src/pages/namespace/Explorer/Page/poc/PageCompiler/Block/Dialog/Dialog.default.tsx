import { DialogBaseComponent, DialogProps } from "./DialogBase";

import { EditModeDialog } from "../../../BlockEditor/PageCompiler/EditorDialog";
import { usePageStateContext } from "../../context/pageCompilerContext";

export const Dialog = (props: DialogProps) => {
  const { mode } = usePageStateContext();

  if (mode === "live") {
    return <DialogBaseComponent {...props} />;
  }

  return <EditModeDialog {...props} />;
};
