import { Blocks, Settings } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/design/Tabs";

import { BlockForm } from "..";
import { DragablePaletteItem } from "~/design/DragAndDrop/Draggable";
import { useBlockTypes } from "../../PageCompiler/context/utils/useBlockTypes";
import { usePageEditorPanel } from "../EditorPanelProvider";
import { useTranslation } from "react-i18next";

export const EditorPanel = () => {
  const { panel } = usePageEditorPanel();
  const { t } = useTranslation();

  const rootLevel = [1];
  const allowedBlockTypes = useBlockTypes(rootLevel);

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
              {allowedBlockTypes.map((type, index) => (
                <DragablePaletteItem
                  key={index}
                  payload={{ type: "add", blockType: type.type }}
                  icon={type.icon}
                >
                  {type.label}
                </DragablePaletteItem>
              ))}
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
