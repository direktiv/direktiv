import {
  AllBlocksType,
  InlineBlocksType,
  inlineBlockTypes,
} from "../../../schema/blocks";

import { BlockPathType } from "../../Block";
import { useBlockTypes } from "./useBlockTypes";
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
  const blockTypes = useBlockTypes();

  const createBlock = (type: AllBlocksType["type"], path: BlockPathType) => {
    const matchingBlockType = blockTypes.find((block) => block.type === type);

    if (!matchingBlockType) {
      throw new Error(`${type} is not implemented yet`);
    }

    if (inlineBlockTypes.has(type as InlineBlocksType)) {
      return addBlock(path, matchingBlockType.defaultValues, true);
    }

    setPanel({
      action: "create",
      block: matchingBlockType.defaultValues,
      path,
    });
  };

  return createBlock;
};
