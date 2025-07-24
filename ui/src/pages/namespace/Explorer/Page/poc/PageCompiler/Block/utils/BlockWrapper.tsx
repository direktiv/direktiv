import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react";
import {
  findAncestor,
  incrementPath,
  pathIsDescendant,
  pathsEqual,
} from "../../context/utils";
import {
  usePage,
  usePageStateContext,
} from "../../context/pageCompilerContext";

import Badge from "~/design/Badge";
import { BlockPathType } from "..";
import { BlockType } from "../../../schema/blocks";
import { DragPayloadSchemaType } from "~/design/DragAndDrop/schema";
import { Dropzone } from "~/design/DragAndDrop/Dropzone";
import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { SortableItem } from "~/design/DragAndDrop/Draggable";
import { twMergeClsx } from "~/util/helpers";
import { useBlockTypes } from "../../context/utils/useBlockTypes";
import { usePageEditorPanel } from "../../../BlockEditor/EditorPanelProvider";
import { useTranslation } from "react-i18next";

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
  const { getAllowedTypes } = useBlockTypes();
  const containerRef = useRef<HTMLDivElement>(null);

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

  const isFocused = panel?.action && pathsEqual(panel.path, blockPath);

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
        dialog: null,
        block: parentDialog.block,
        path: parentDialog.path,
      });
    }

    if (isFocused) {
      return setPanel(null);
    }

    return setPanel({
      action: "edit",
      dialog: null,
      block,
      path: blockPath,
    });
  };

  const nextSilblingPath = incrementPath(blockPath);

  const enableDropZone = (payload: DragPayloadSchemaType | null) => {
    if (panel?.dialog && !pathIsDescendant(blockPath, panel.dialog)) {
      return false;
    }

    const allowedTypes = getAllowedTypes(nextSilblingPath);
    if (!allowedTypes.some((config) => config.type === payload?.blockType)) {
      return false;
    }

    if (payload?.type === "move") {
      // don't show a dropzone for neighboring blocks
      if (
        pathsEqual(payload.originPath, nextSilblingPath) ||
        pathsEqual(payload.originPath, blockPath)
      ) {
        return false;
      }
    }
    return true;
  };

  return (
    <>
      <SortableItem
        payload={{
          type: "move",
          block,
          blockType: block.type,
          originPath: blockPath,
        }}
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
      </SortableItem>
      <Dropzone
        payload={{ targetPath: nextSilblingPath }}
        enable={enableDropZone}
      />
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
