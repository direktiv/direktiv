import {} from "../../PageCompiler/Block/Dialog/Dialog.default";

import {
  DialogBaseComponent,
  DialogProps,
} from "../../PageCompiler/Block/Dialog/DialogBase";

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
