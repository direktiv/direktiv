import { pathIsDescendant, pathsEqual } from "../../context/utils";

import { BlockPathType } from "..";
import { DragPayloadSchemaType } from "~/design/DragAndDrop/schema";
import { useBlockTypes } from "../../context/utils/useBlockTypes";
import { usePageEditorPanel } from "../../../BlockEditor/EditorPanelProvider";

export const useValidateDropzone = () => {
  const { panel } = usePageEditorPanel();
  const { getAllowedTypes } = useBlockTypes();

  const enable = (
    payload: DragPayloadSchemaType | null,
    targetPath: BlockPathType
  ) => {
    if (panel?.dialog && !pathIsDescendant(targetPath, panel.dialog)) {
      return false;
    }

    const allowedTypes = getAllowedTypes(targetPath);
    if (!allowedTypes.some((config) => config.type === payload?.blockType)) {
      return false;
    }

    if (payload?.type === "move") {
      // don't show a dropzone for neighboring blocks
      if (pathsEqual(payload.originPath, targetPath)) {
        return false;
      }
    }
    return true;
  };

  return enable;
};
