import {
  incrementPath,
  pathIsDescendant,
  pathsEqual,
} from "../../context/utils";

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
    if (!payload) {
      return false;
    }
    if (panel?.dialog && !pathIsDescendant(targetPath, panel.dialog)) {
      return false;
    }

    const allowedTypes = getAllowedTypes(targetPath);

    const blockType =
      payload.type === "move" ? payload.block.type : payload.blockType;

    if (!allowedTypes.some((config) => config.type === blockType)) {
      return false;
    }

    if (payload.type === "move") {
      // don't show a dropzone for same or neighboring index
      if (
        pathsEqual(payload.originPath, targetPath) ||
        pathsEqual(incrementPath(payload.originPath), targetPath)
      ) {
        return false;
      }
    }
    return true;
  };

  return enable;
};
