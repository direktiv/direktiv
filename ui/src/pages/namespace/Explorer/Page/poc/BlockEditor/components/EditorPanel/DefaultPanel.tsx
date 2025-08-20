import { Blocks, Settings } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/design/Tabs";

import { DraggablePaletteItem } from "~/design/DragAndDrop/Draggable";
import { useBlockTypes } from "../../../PageCompiler/context/utils/useBlockTypes";
import { useTranslation } from "react-i18next";

export const DefaultPanel = () => {
  const { t } = useTranslation();

  const { blockTypes } = useBlockTypes();

  return (
    <div className="h-[300px] overflow-y-clip border-b-2 border-gray-4 dark:border-gray-dark-4 sm:h-[calc(100vh-230px)] sm:border-b-0 sm:border-r-2">
      <Tabs defaultValue="addBlock">
        <TabsList className="w-full rounded-none border-b border-gray-5 bg-gray-1 p-5 pb-0 dark:border-gray-dark-5 dark:bg-gray-dark-1">
          <TabsTrigger className="w-full" value="addBlock">
            <Blocks size={16} />
            {t("direktivPage.blockEditor.generic.addBlockTab")}
          </TabsTrigger>
          <TabsTrigger className="w-full" value="settings">
            <Settings size={16} />
            {t("direktivPage.blockEditor.generic.settingsTab")}
          </TabsTrigger>
        </TabsList>
        <TabsContent className="p-3" value="addBlock" asChild>
          <div className="grid grid-cols-2 gap-2">
            {blockTypes.map((type, index) => (
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
    </div>
  );
};
