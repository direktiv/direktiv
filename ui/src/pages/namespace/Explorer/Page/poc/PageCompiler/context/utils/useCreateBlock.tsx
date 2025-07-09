import {
  AllBlocksType,
  InlineBlocksType,
  inlineBlockTypes,
} from "../../../schema/blocks";

import { BlockPathType } from "../../Block";
import { getBlockTemplate } from ".";
import { usePageEditor } from "../pageCompilerContext";
import { usePageEditorPanel } from "../../../BlockEditor/EditorPanelProvider";

/**
 * This hook returns createBlock(), which opens the editor form for
 * blocks such as text blocks, or just adds an inline block to the page
 * (e.g., cards or columns, where no form is required).
 */
export const useCreateBlock = () => {
  const { addBlock } = usePageEditor();
  const { setPanel } = usePageEditorPanel();

  const createBlock = (type: AllBlocksType["type"], path: BlockPathType) => {
    if (inlineBlockTypes.has(type as InlineBlocksType)) {
      return addBlock(path, getBlockTemplate(type), true);
    }
    setPanel({
      action: "create",
      block: getBlockTemplate(type),
      path,
    });
  };

  return createBlock;
};
