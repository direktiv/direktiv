import { DialogProps } from "./DialogBase";
import { EditModeDialog } from "../../../BlockEditor/PageCompiler/EditorDialog";
import { Dialog as LiveModeDialog } from "./Dialog.pagesapp";
import { usePageStateContext } from "../../context/pageCompilerContext";

export const Dialog = (props: DialogProps) => {
  const { mode } = usePageStateContext();

  if (mode === "live") {
    return <LiveModeDialog {...props} />;
  }

  return <EditModeDialog {...props} />;
};
