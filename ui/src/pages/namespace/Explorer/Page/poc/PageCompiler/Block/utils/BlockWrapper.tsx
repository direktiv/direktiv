import {
  LocalVariables,
  useVariablesContext,
} from "../../primitives/Variable/VariableContext";
import { ReactElement, useEffect, useMemo, useRef, useState } from "react";
import {
  findAncestor,
  incrementPath,
  isFirstChildPath,
  pathsEqual,
} from "../../context/utils";
import {
  usePage,
  usePageStateContext,
} from "../../context/pageCompilerContext";

import { BlockPathType } from "..";
import { BlockSuspenseBoundary } from "./SuspenseBoundary";
import { BlockType } from "../../../schema/blocks";
import { Dropzone } from "~/design/DragAndDrop/Dropzone";
import { SortableItem } from "~/design/DragAndDrop/Draggable";
import { isEmptyContainerBlock } from "./useIsInvisbleBlock";
import { twMergeClsx } from "~/util/helpers";
import { useDndContext } from "@dnd-kit/core";
import { usePageEditorPanel } from "../../../BlockEditor/EditorPanelProvider";
import { useValidateDropzone } from "./useValidateDropzone";

type BlockWrapperProps = {
  blockPath: BlockPathType;
  block: BlockType;
  children: (register?: (vars: LocalVariables) => void) => ReactElement;
};

const EditorBlockWrapper = ({
  block,
  blockPath,
  children,
}: BlockWrapperProps) => {
  const page = usePage();
  const { panel, setPanel } = usePageEditorPanel();
  const [isHovered, setIsHovered] = useState(false);
  const contextVariables = useVariablesContext();
  const [localVariables, setLocalVariables] = useState<LocalVariables>({
    this: {},
  });
  const validateDropzone = useValidateDropzone();
  const containerRef = useRef<HTMLDivElement>(null);
  const dndContext = useDndContext();
  const isDragging = !!dndContext.active;

  const variables = useMemo(
    () =>
      block.type === "form"
        ? { ...contextVariables, ...localVariables }
        : contextVariables,
    [block, contextVariables, localVariables]
  );

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (containerRef.current) {
        const allBlockWrapper = Array.from(
          document.querySelectorAll("[data-block-wrapper]")
        ).filter((element) => element.contains(e.target as Node));

        const deepestChildren = allBlockWrapper.at(-1);
        setIsHovered(containerRef.current === deepestChildren);
      }
    };

    document.addEventListener("mousemove", handleMouseMove);
    return () => document.removeEventListener("mousemove", handleMouseMove);
  }, []);

  const dropzonePayload = useMemo(
    () => ({
      targetPath: incrementPath(blockPath),
      variables,
    }),
    [blockPath, variables]
  );

  const isFocused = !!(panel?.action && pathsEqual(panel.path, blockPath));

  const handleClickBlock = (event: React.MouseEvent<HTMLDivElement>) => {
    event.stopPropagation();

    const parentDialog = findAncestor({
      page,
      path: blockPath,
      match: (block) => block.type === "dialog",
    });

    // if block losing focus is in a Dialog, focus the Dialog.
    if (isFocused && parentDialog) {
      return setPanel({
        action: "edit",
        block: parentDialog.block,
        path: parentDialog.path,
        variables,
      });
    }

    // if focused block is clicked, unfocus
    if (isFocused) {
      return setPanel(null);
    }

    // if unfocused block is clicked, focus it
    return setPanel({
      action: "edit",
      block,
      path: blockPath,
      variables,
    });
  };

  const showDragHandle = isHovered || isFocused;

  const emptyContainerBlock = isEmptyContainerBlock(block);

  return (
    <>
      {isFirstChildPath(blockPath) && (
        <Dropzone
          payload={{ ...dropzonePayload, targetPath: blockPath }}
          validate={validateDropzone}
        />
      )}
      <SortableItem
        payload={{
          type: "move",
          block,
          originPath: blockPath,
        }}
        blockPath={blockPath}
        isFocused={isFocused}
        className={twMergeClsx(showDragHandle ? "visible" : "invisible")}
      >
        <div
          ref={containerRef}
          className={twMergeClsx(
            "relative isolate my-3 rounded bg-white outline-offset-4 dark:bg-black",
            emptyContainerBlock &&
              "border-2 border-dashed border-gray-4 dark:border-gray-dark-4",
            isHovered &&
              !isDragging &&
              "bg-gray-2 outline outline-2 outline-gray-4 dark:bg-gray-dark-2 dark:outline-gray-dark-4",
            isFocused &&
              "border-gray-8 outline outline-2 outline-gray-8 dark:border-gray-8 dark:outline-gray-dark-8",
            isDragging && "outline outline-gray-7 dark:outline-gray-dark-7"
          )}
          data-block-wrapper
          onClick={handleClickBlock}
        >
          <BlockSuspenseBoundary>
            {block.type === "form" ? children(setLocalVariables) : children()}
          </BlockSuspenseBoundary>
        </div>
      </SortableItem>
      <Dropzone payload={dropzonePayload} validate={validateDropzone} />
    </>
  );
};

const VisitorBlockWrapper = ({ children }: BlockWrapperProps) => (
  <BlockSuspenseBoundary>
    <div className="my-3">{children()}</div>
  </BlockSuspenseBoundary>
);

export const BlockWrapper = (props: BlockWrapperProps) => {
  const { mode } = usePageStateContext();

  if (mode === "edit") {
    return <EditorBlockWrapper {...props} />;
  }

  return <VisitorBlockWrapper {...props} />;
};
