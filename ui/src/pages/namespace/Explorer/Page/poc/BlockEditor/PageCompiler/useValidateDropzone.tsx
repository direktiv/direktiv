import {
  incrementPath,
  pathIsDescendant,
  pathsEqual,
} from "../../PageCompiler/context/utils";

import { BlockPathType } from "../../PageCompiler/Block";
import { DragPayloadSchemaType } from "~/design/DragAndDrop/schema";
import { DropzoneStatus } from "~/design/DragAndDrop/Dropzone";
import { useAllowedBlockTypes } from "../../PageCompiler/context/utils/useBlockTypes";
import { useCallback } from "react";
import { usePageEditorPanel } from "../EditorPanelProvider";

export const useValidateDropzone = () => {
  const { dialog } = usePageEditorPanel();
  const getAllowedTypes = useAllowedBlockTypes();

  const enable = useCallback(
    (
      payload: DragPayloadSchemaType | null,
      targetPath: BlockPathType
    ): DropzoneStatus => {
      if (!payload) {
        return "hidden";
      }

      if (dialog && !pathIsDescendant(targetPath, dialog)) {
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
    },
    [dialog, getAllowedTypes]
  );

  return enable;
};
