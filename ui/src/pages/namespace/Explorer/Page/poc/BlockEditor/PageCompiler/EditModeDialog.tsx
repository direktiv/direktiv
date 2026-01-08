import {
  DialogBaseComponent,
  DialogProps,
} from "../../PageCompiler/Block/Dialog";

import { usePageEditorPanel } from "../../BlockEditor/EditorPanelProvider";

export const EditModeDialog = (props: DialogProps) => {
  const { setDialog } = usePageEditorPanel();

  return (
    <DialogBaseComponent
      {...props}
      onOpenChange={(open) => setDialog(open ? props.blockPath : null)}
    />
  );
};
