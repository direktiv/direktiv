import { Dialog, DialogContent } from "~/design/Dialog";
import { PropsWithChildren, createContext, useContext, useState } from "react";
import {
  usePageEditor,
  usePageStateContext,
} from "../PageCompiler/context/pageCompilerContext";

import { ActionPanel } from "./components/EditorPanel/ActionPanel";
import { BlockDeleteForm } from "./components/Delete";
import { BlockPathType } from "../PageCompiler/Block";
import { BlockType } from "../schema/blocks";
import { ContextVariables } from "../PageCompiler/primitives/Variable/VariableContext";
import { DefaultPanel } from "./components/EditorPanel/DefaultPanel";
import { DndContext } from "~/design/DragAndDrop";
import { DragAndDropPayloadSchemaType } from "~/design/DragAndDrop/schema";
import { LocalDialogContainer } from "~/design/LocalDialog/container";
import { getBlockConfig } from "../PageCompiler/context/utils/useBlockTypes";

export type EditorPanelAction = {
  action: "create" | "edit" | "delete";
  block: BlockType;
  path: BlockPathType;
};

type EditorPanelState = null | EditorPanelAction;

type EditorDialogState = null | BlockPathType;

type EditorPanelContextType = {
  panel: EditorPanelState;
  setPanel: React.Dispatch<React.SetStateAction<EditorPanelState>>;
  dialog: EditorDialogState;
  setDialog: React.Dispatch<React.SetStateAction<EditorDialogState>>;
  variables: ContextVariables;
  setVariables: React.Dispatch<React.SetStateAction<ContextVariables>>;
};

const EditorPanelContext = createContext<EditorPanelContextType | null>(null);

const PagePreviewContainer = ({ children }: PropsWithChildren) => (
  <div className="grow sm:h-[calc(100vh-230px)] sm:overflow-y-scroll">
    <LocalDialogContainer className="h-full min-w-0 flex-1 px-3 py-5">
      <div className="mx-auto max-w-screen-lg">{children}</div>
    </LocalDialogContainer>
  </div>
);

export const EditorPanelLayoutProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const { addBlock, deleteBlock, moveBlock } = usePageEditor();
  const [panel, setPanel] = useState<EditorPanelState>(null);
  const [dialog, setDialog] = useState<EditorDialogState>(null);
  const [variables, setVariables] = useState<ContextVariables>({
    loop: {},
    query: {},
  });
  const { mode } = usePageStateContext();

  const createBlock = (type: BlockType["type"], path: BlockPathType) => {
    const blockConfig = getBlockConfig(type);

    if (!blockConfig) throw new Error(`No blockConfig found for ${type}`);

    if (!blockConfig.formComponent) {
      return addBlock(path, blockConfig.defaultValues);
    }
    setPanel({
      action: "create",
      block: blockConfig.defaultValues,
      path,
    });
  };

  const onDrop = (payload: DragAndDropPayloadSchemaType) => {
    const { drag, drop } = payload;
    if (drag.type === "add") {
      createBlock(drag.blockType, [...drop.targetPath]);
    }
    if (drag.type === "move") {
      moveBlock(drag.originPath, drop.targetPath, drag.block);
      setPanel({
        action: "edit",
        path: drop.targetPath,
        block: drag.block,
      });
    }
  };

  if (mode === "edit") {
    return (
      <DndContext onDrop={onDrop} onDrag={() => setPanel(null)}>
        <EditorPanelContext.Provider
          value={{
            panel,
            setPanel,
            dialog,
            setDialog,
            variables,
            setVariables,
          }}
        >
          <div className="grow sm:grid sm:grid-cols-[350px_1fr]">
            {panel?.action ? <ActionPanel panel={panel} /> : <DefaultPanel />}
            <PagePreviewContainer>{children}</PagePreviewContainer>
          </div>
          <Dialog open={panel && panel.action === "delete" ? true : false}>
            <DialogContent>
              {!!panel && panel.action === "delete" && (
                <BlockDeleteForm
                  path={panel.path}
                  onSubmit={() => {
                    deleteBlock(panel.path);
                    setPanel(null);
                  }}
                  onCancel={() => {
                    setPanel({ ...panel, action: "edit" });
                  }}
                />
              )}
            </DialogContent>
          </Dialog>
        </EditorPanelContext.Provider>
      </DndContext>
    );
  }

  return <PagePreviewContainer>{children}</PagePreviewContainer>;
};

export const usePageEditorPanel = () => {
  const context = useContext(EditorPanelContext);
  if (!context)
    throw new Error(
      "usePageEditorPanel must be used within EditorPanelLayoutProvider"
    );
  return context;
};
