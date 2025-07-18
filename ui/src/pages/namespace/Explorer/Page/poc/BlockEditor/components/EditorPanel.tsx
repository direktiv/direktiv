import { Blocks, Settings } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/design/Tabs";

import { BlockForm } from "..";
import { Card } from "~/design/Card";
import { DraggableCreateElement } from "~/design/DragAndDropEditor/DraggableElement";
import { useBlockTypes } from "../../PageCompiler/context/utils/useBlockTypes";
import { usePageEditorPanel } from "../EditorPanelProvider";
import { useTranslation } from "react-i18next";

export const EditorPanel = () => {
  const { panel } = usePageEditorPanel();
  const { t } = useTranslation();

  const path = panel?.path ?? [1];

  const types = useBlockTypes(path);

  if (!panel) {
    return (
      <div>
        <Tabs defaultValue="addBlock">
          <TabsList variant="boxed">
            <TabsTrigger variant="boxed" value="addBlock">
              <Blocks size={16} />
              {t("direktivPage.blockEditor.generic.addBlockTab")}
            </TabsTrigger>
            <TabsTrigger variant="boxed" value="settings">
              <Settings size={16} />
              {t("direktivPage.blockEditor.generic.settingsTab")}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="addBlock" asChild>
            <div className="relative flex-col-reverse overflow-visible">
              {types.map((type, index) => {
                const Icon = type.icon;
                return (
                  <DraggableCreateElement
                    key={type.label}
                    id={index}
                    type={type.type}
                  >
                    <Card className="z-50 m-4 flex justify-center bg-gray-2 p-4 text-sm text-black dark:bg-gray-dark-2 dark:text-white">
                      <Icon size={16} className="mr-4" /> {type.label}
                    </Card>
                  </DraggableCreateElement>
                );
              })}
            </div>
          </TabsContent>
          <TabsContent value="settings" asChild></TabsContent>
        </Tabs>
      </div>
    );
  }

  return (
    <BlockForm action={panel.action} path={panel.path} block={panel.block} />
  );
};
