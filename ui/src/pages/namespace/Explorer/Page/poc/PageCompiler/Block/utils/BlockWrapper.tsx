import { Dialog, DialogContent } from "~/design/Dialog";
import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react";
import { getPlaceholderBlock, pathsEqual } from "../../context/utils";

import { AllBlocksType } from "../../../schema/blocks";
import Badge from "~/design/Badge";
import { BlockContextMenu } from "../../../BlockEditor/components/ContextMenu";
import { BlockDeleteForm } from "../../../BlockEditor/components/Delete";
import { BlockForm } from "../../../BlockEditor";
import { BlockPathType } from "..";
import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { SelectBlockType } from "../../../BlockEditor/components/SelectType";
import { twMergeClsx } from "~/util/helpers";
import { useBlockDialog } from "../../../BlockEditor/useBlockDialog";
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
  const { dialog, setDialog } = useBlockDialog();
  const { deleteBlock } = usePageEditor();

  /**
   * This handler is only used for closing the dialog. For opening a dialog,
   * we add custom onClick events to the trigger buttons.
   */
  const handleOnOpenChange = (open: boolean) => {
    if (open === false) {
      setDialog(null);
    }
  };

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
      <div
        ref={containerRef}
        className={twMergeClsx(
          mode === "edit" &&
            "relative rounded-md border-2 border-dashed border-gray-4 bg-white p-3 dark:border-gray-dark-4 dark:bg-black",
          isHovered &&
            mode === "edit" &&
            "border-solid bg-gray-2 dark:bg-gray-dark-2",
          isFocused &&
            mode === "edit" &&
            "border-solid border-gray-8 dark:border-gray-10"
        )}
        data-block-wrapper
        onClick={handleClickBlock}
      >
        {mode === "edit" && (isHovered || isFocused) && (
          <Badge className="absolute z-30 -m-6" variant="secondary">
            <b>{block.type}</b>
            {blockPath.join(".")}
          </Badge>
        )}
        {mode === "edit" && isFocused && (
          <div onClick={(event) => event.stopPropagation()}>
            <Dialog open={!!dialog} onOpenChange={handleOnOpenChange}>
              <div className="absolute right-1 top-1 z-30">
                <BlockContextMenu
                  onEdit={() =>
                    setDialog({ action: "edit", blockType: block.type })
                  }
                  onDelete={() =>
                    setDialog({ action: "delete", blockType: block.type })
                  }
                />
              </div>
              <SelectBlockType
                onSelect={(type) => {
                  setDialog({ action: "create", blockType: type });
                }}
              />
              {dialog !== null && (
                <DialogContent className="z-50">
                  {dialog.action === "edit" && (
                    <BlockForm
                      block={block}
                      action={dialog.action}
                      path={blockPath}
                    />
                  )}
                  {dialog.action === "create" && (
                    <BlockForm
                      block={getPlaceholderBlock(dialog.blockType)}
                      action={dialog.action}
                      path={blockPath}
                    />
                  )}
                  {dialog.action === "delete" && (
                    <BlockDeleteForm
                      type={block.type}
                      action={dialog.action}
                      path={blockPath}
                      onSubmit={(path) => deleteBlock(path)}
                    />
                  )}
                </DialogContent>
              )}
            </Dialog>
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
    </>
  );
};
