import {
  incrementPath,
  pathIsDescendant,
  pathsEqual,
} from "../../context/utils";

import { BlockPathType } from "..";
import { DragPayloadSchemaType } from "~/design/DragAndDrop/schema";
import { DropzoneStatus } from "~/design/DragAndDrop/Dropzone";
import { useBlockTypes } from "../../context/utils/useBlockTypes";
import { usePageEditorPanel } from "../../../BlockEditor/EditorPanelProvider";

export const useValidateDropzone = () => {
  const { panel } = usePageEditorPanel();
  const { getAllowedTypes } = useBlockTypes();

  const enable = (
    payload: DragPayloadSchemaType | null,
    targetPath: BlockPathType
  ): DropzoneStatus => {
    if (!payload) {
      return "hidden";
    }
    if (panel?.dialog && !pathIsDescendant(targetPath, panel.dialog)) {
      return "hidden";
    }

    if (payload.type === "move") {
      // don't show a dropzone for same, neighboring or nested index
      if (
        pathsEqual(payload.originPath, targetPath) ||
        pathsEqual(incrementPath(payload.originPath), targetPath) ||
        pathIsDescendant(targetPath, payload.originPath)
      ) {
        return "hidden";
      }
    }

    const allowedTypes = getAllowedTypes(targetPath);

    const blockType =
      payload.type === "move" ? payload.block.type : payload.blockType;

    if (!allowedTypes.some((config) => config.type === blockType)) {
      return "forbidden";
    }

    return "allowed";
  };

  return enable;
};
