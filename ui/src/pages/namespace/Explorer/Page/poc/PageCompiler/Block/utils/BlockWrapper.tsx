import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { findAncestor, incrementPath, pathsEqual } from "../../context/utils";
import {
  usePage,
  usePageStateContext,
} from "../../context/pageCompilerContext";

import { BlockPathType } from "..";
import { BlockType } from "../../../schema/blocks";
import { Dropzone } from "~/design/DragAndDrop/Dropzone";
import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { SortableItem } from "~/design/DragAndDrop/Draggable";
import { twMergeClsx } from "~/util/helpers";
import { useDndContext } from "@dnd-kit/core";
import { usePageEditorPanel } from "../../../BlockEditor/EditorPanelProvider";
import { useTranslation } from "react-i18next";
import { useValidateDropzone } from "./useValidateDropzone";

type BlockWrapperProps = PropsWithChildren<{
  blockPath: BlockPathType;
  block: BlockType;
}>;

const EditorBlockWrapper = ({
  block,
  blockPath,
  children,
}: BlockWrapperProps) => {
  const { t } = useTranslation();
  const page = usePage();
  const { panel, setPanel } = usePageEditorPanel();
  const [isHovered, setIsHovered] = useState(false);
  const validateDropzone = useValidateDropzone();
  const containerRef = useRef<HTMLDivElement>(null);
  const dndContext = useDndContext();
  const isDragging = !!dndContext.active;

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
    }),
    [blockPath]
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
      });
    }

    // if focused block is clicked, unfocus
    if (isFocused) {
      return setPanel(null);
    }

    return setPanel({
      action: "edit",
      block,
      path: blockPath,
    });
  };

  const showDragHandle = isHovered || isFocused;

  const emptyColumnsBlock =
    block.type === "columns" &&
    block.blocks[0] &&
    block.blocks[1] &&
    block.blocks[0].blocks.length === 0 &&
    block.blocks[1].blocks.length === 0;

  const emptyQueryBlock =
    block.type === "query-provider" && block.blocks.length === 0;

  const emptyBlock = emptyColumnsBlock || emptyQueryBlock;

  return (
    <>
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
            emptyBlock &&
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
          <Suspense fallback={<Loading />}>
            <ErrorBoundary
              fallbackRender={({ error }) => (
                <ParsingError title={t("direktivPage.error.genericError")}>
                  {error.message}
                </ParsingError>
              )}
            >
              {children}
            </ErrorBoundary>
          </Suspense>
        </div>
      </SortableItem>
      <Dropzone payload={dropzonePayload} validate={validateDropzone} />
    </>
  );
};

const VisitorBlockWrapper = ({ children }: BlockWrapperProps) => {
  const { t } = useTranslation();

  return (
    <Suspense fallback={<Loading />}>
      <ErrorBoundary
        fallbackRender={({ error }) => (
          <ParsingError title={t("direktivPage.error.genericError")}>
            {error.message}
          </ParsingError>
        )}
      >
        <div className="my-3">{children}</div>
      </ErrorBoundary>
    </Suspense>
  );
};

export const BlockWrapper = (props: BlockWrapperProps) => {
  const { mode } = usePageStateContext();

  if (mode === "edit") {
    return <EditorBlockWrapper {...props} />;
  }

  return <VisitorBlockWrapper {...props} />;
};
