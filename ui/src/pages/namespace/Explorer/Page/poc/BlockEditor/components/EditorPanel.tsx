import { Blocks, Settings } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/design/Tabs";

import { AllBlocksType } from "../../schema/blocks";
import { BlockForm } from "..";
import { Card } from "~/design/Card";
import { DraggableCreateElement } from "~/design/DragAndDropEditor/DraggableElement";
import { useBlockTypes } from "../../PageCompiler/context/utils/useBlockTypes";
import { usePageEditorPanel } from "../EditorPanelProvider";

export const EditorPanel = () => {
  const { panel } = usePageEditorPanel();

  const path = panel?.path ?? [1];

  const types = useBlockTypes(path);

  const emptyBlock: AllBlocksType = { type: "button", label: "ok" };

  if (!panel) {
    return (
      <div>
        <Tabs
          defaultValue="general"
          className="z-50 w-full overflow-visible border-red-300"
        >
          <TabsList variant="boxed">
            <TabsTrigger variant="boxed" value="general">
              <Settings size={16} /> General Settings
            </TabsTrigger>
            <TabsTrigger variant="boxed" value="blockcollection">
              <Blocks size={16} />
              Add Block
            </TabsTrigger>
          </TabsList>
          <TabsContent value="general" asChild>
            <Card
              className="row flex bg-gray-2 p-4 text-sm dark:bg-gray-dark-2"
              noShadow
            >
              settings
            </Card>
          </TabsContent>
          <TabsContent value="blockcollection" asChild>
            <div className="relative flex-col-reverse overflow-visible">
              {types.map((type) => {
                const Icon = type.icon;
                return (
                  <DraggableCreateElement
                    key={type.label}
                    id={String(type.label)}
                    element={emptyBlock}
                    blockPath={null}
                  >
                    <Card className="z-50 m-4 flex justify-center bg-gray-2 p-4 text-sm text-black dark:bg-gray-dark-2">
                      <Icon size={16} className="mr-4" /> {type.label}
                    </Card>
                  </DraggableCreateElement>
                );
              })}
            </div>
          </TabsContent>
        </Tabs>
      </div>
    );
  }

  return (
    <BlockForm action={panel.action} path={panel.path} block={panel.block} />
  );
};
