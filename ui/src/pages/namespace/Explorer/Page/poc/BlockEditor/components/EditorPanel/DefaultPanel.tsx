import { Blocks } from "lucide-react";
import { DraggablePaletteItem } from "~/design/DragAndDrop/Draggable";
import { PanelContainer } from "./PanelContainer";
import { blockTypes } from "../../../PageCompiler/context/utils/useBlockTypes";
import { useTranslation } from "react-i18next";

export const DefaultPanel = () => {
  const { t } = useTranslation();

  return (
    <PanelContainer className="overflow-y-clip">
      <div className="w-full rounded rounded-b-none border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
        <h3 className="flex grow gap-x-2 font-bold">
          <Blocks className="h-5" />
          {t("direktivPage.blockEditor.generic.addBlockTab")}
        </h3>
      </div>
      <div className="grid grid-cols-2 gap-2 p-5">
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
    </PanelContainer>
  );
};
