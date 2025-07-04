import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react";
import { pathToId, pathsEqual } from "../../context/utils";

import { AllBlocksType } from "../../../schema/blocks";
import Badge from "~/design/Badge";
import { BlockContextMenu } from "../../../BlockEditor/components/ContextMenu";
import { BlockPathType } from "..";
import { DraggableElement } from "~/design/DragAndDropEditor/DraggableElement";
import { DroppableSeparator } from "~/design/DragAndDropEditor/DroppableSeparator";
import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { SelectBlockType } from "../../../BlockEditor/components/SelectType";
import { twMergeClsx } from "~/util/helpers";
import { useBlockDialog } from "../../../BlockEditor/BlockDialogProvider";
import { useCreateBlock } from "../../context/utils/useCreateBlock";
import { usePageEditor } from "../../context/pageCompilerContext";
import { useTranslation } from "react-i18next";

type BlockWrapperProps = PropsWithChildren<{
  blockPath: BlockPathType;
  block: AllBlocksType;
}>;

export const BlockWrapper = ({
  block,
  blockPath,
  children,
}: BlockWrapperProps) => {
  const { t } = useTranslation();
  const { mode, focus, setFocus } = usePageEditor();
  const [isHovered, setIsHovered] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const { setDialog } = useBlockDialog();
  const createBlock = useCreateBlock();

  useEffect(() => {
    if (mode !== "edit") {
      return;
    }

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
  }, [mode]);

  const handleClickBlock = (event: React.MouseEvent<HTMLDivElement>) => {
    event.stopPropagation();
    if (mode !== "edit") {
      return;
    }
    return setFocus(blockPath);
  };

  const isFocused = focus && pathsEqual(focus, blockPath);

  return (
    <>
      {mode === "edit" ? (
        <>
          <DroppableSeparator
            id={pathToId(blockPath)}
            blockPath={blockPath}
            position="before"
          ></DroppableSeparator>
          <DraggableElement
            blockPath={blockPath}
            element={block}
            id={pathToId(blockPath)}
          >
            <div
              ref={containerRef}
              className={twMergeClsx(
                "relative rounded-md rounded-s-none border-2 border-dashed border-gray-4 bg-white p-0 dark:border-gray-dark-4 dark:bg-black",
                isHovered && "border-solid bg-gray-2 dark:bg-gray-dark-2",
                isFocused && "border-solid border-gray-8 dark:border-gray-8"
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
              {isFocused && (
                <div onClick={(event) => event.stopPropagation()}>
                  <div className="absolute right-1 top-1 z-30">
                    <BlockContextMenu
                      onEdit={() =>
                        setDialog({
                          action: "edit",
                          block,
                          path: blockPath,
                        })
                      }
                      onDelete={() =>
                        setDialog({
                          action: "delete",
                          block,
                          path: blockPath,
                        })
                      }
                    />
                  </div>

                  <SelectBlockType
                    path={blockPath}
                    onSelect={(type) => createBlock(type, blockPath)}
                  />
                </div>
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
      ) : (
        <>
          <div ref={containerRef} data-block-wrapper onClick={handleClickBlock}>
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
        </>
      )}
    </>
  );
};
