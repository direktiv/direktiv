import { Blocks, Settings } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/design/Tabs";

import { BlockForm } from "..";
import { DraggablePaletteItem } from "~/design/DragAndDrop/Draggable";
import { useBlockTypes } from "../../PageCompiler/context/utils/useBlockTypes";
import { usePageEditorPanel } from "../EditorPanelProvider";
import { useTranslation } from "react-i18next";

export const EditorPanel = () => {
  const { panel } = usePageEditorPanel();
  const { t } = useTranslation();

  const rootLevel = [1];
  const { getAllowedTypes } = useBlockTypes();
  const allowedBlockTypes = getAllowedTypes(rootLevel);

  if (!panel) {
    return (
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
          <div className="grid grid-cols-3 gap-2 overflow-visible lg:grid-cols-1">
            {allowedBlockTypes.map((type, index) => (
              <DraggablePaletteItem
                key={index}
                payload={{ type: "add", blockType: type.type }}
                icon={type.icon}
              >
                {type.label}
              </DraggablePaletteItem>
            ))}
          </div>
        </TabsContent>
        <TabsContent value="settings" asChild></TabsContent>
      </Tabs>
    );
  }

  return (
    <BlockForm action={panel.action} path={panel.path} block={panel.block} />
  );
};
