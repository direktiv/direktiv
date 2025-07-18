import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react";
import { findAncestor, pathToId, pathsEqual } from "../../context/utils";
import {
  usePage,
  usePageStateContext,
} from "../../context/pageCompilerContext";

import { AllBlocksType } from "../../../schema/blocks";
import Badge from "~/design/Badge";
import { BlockPathType } from "..";
import { DraggableElement } from "~/design/DragAndDropEditor/DraggableElement";
import { DroppableSeparator } from "~/design/DragAndDropEditor/DroppableSeparator";
import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { twMergeClsx } from "~/util/helpers";
import { useCreateBlock } from "../../context/utils/useCreateBlock";
import { usePageEditorPanel } from "../../../BlockEditor/EditorPanelProvider";
import { useTranslation } from "react-i18next";

type BlockWrapperProps = PropsWithChildren<{
  blockPath: BlockPathType;
  block: AllBlocksType;
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
  const containerRef = useRef<HTMLDivElement>(null);
  const createBlock = useCreateBlock();

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

  const isFocused = panel?.path && pathsEqual(panel.path, blockPath);

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

    if (isFocused) {
      return setPanel(null);
    }

    return setPanel({
      action: "edit",
      block,
      path: blockPath,
    });
  };

  return (
    <>
      <DroppableSeparator
        id={pathToId(blockPath)}
        blockPath={blockPath}
        position="before"
        onDrop={(type) => {
          createBlock(type, blockPath);
        }}
      />
      <DraggableElement
        blockPath={blockPath}
        element={block}
        id={pathToId(blockPath)}
      >
        <div
          ref={containerRef}
          className={twMergeClsx(
            "relative isolate z-0 rounded-md rounded-s-none border-2 border-gray-4 bg-white p-0 dark:border-gray-dark-4 dark:bg-black",
            isHovered && "bg-gray-2 dark:bg-gray-dark-2",
            isFocused && "border-gray-8 dark:border-gray-8"
          )}
          data-block-wrapper
          onClick={handleClickBlock}
        >
          {(isHovered || isFocused) && (
            <Badge className="absolute z-30 -m-6" variant="secondary">
              <span className="mr-2">
                <b>{block.type}</b>
              </span>
              {blockPath.join(".")}
            </Badge>
          )}
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
      </DraggableElement>
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
        {children}
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
