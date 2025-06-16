import { Dialog, DialogContent } from "~/design/Dialog";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";
import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react";
import { buttons, getPlaceholderBlock, pathsEqual } from "../../context/utils";

import { AllBlocksType } from "../../../schema/blocks";
import Badge from "~/design/Badge";
import { BlockContextMenu } from "../../../BlockEditor/components/ContextMenu";
import { BlockDeleteForm } from "../../../BlockEditor/components/Delete";
import { BlockForm } from "../../../BlockEditor";
import { BlockPathType } from "..";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { CirclePlus } from "lucide-react";
import { DraggableElement } from "~/design/DragAndDropEditor/DraggableElement";
import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { twMergeClsx } from "~/util/helpers";
import { usePageEditor } from "../../context/pageCompilerContext";
import { useTranslation } from "react-i18next";

type BlockWrapperProps = PropsWithChildren<{
  blockPath: BlockPathType;
  block: AllBlocksType;
}>;

type DialogState = "create" | "edit" | "delete" | null;

export const BlockWrapper = ({
  block,
  blockPath,
  children,
}: BlockWrapperProps) => {
  const { t } = useTranslation();
  const { mode, focus, setFocus } = usePageEditor();
  const [dialog, setDialog] = useState<DialogState>(null);
  const [type, setType] = useState<AllBlocksType["type"]>(block.type);
  const [isHovered, setIsHovered] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

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
      <DraggableElement name={String(blockPath)}>
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
              <span className="mr-2">
                <b>{block.type}</b>
              </span>
              {blockPath.join(".")}
            </Badge>
          )}
          {mode === "edit" && isFocused && (
            <div onClick={(event) => event.stopPropagation()}>
              <Dialog open={!!dialog} onOpenChange={handleOnOpenChange}>
                <div className="absolute right-1 top-1 z-30">
                  <BlockContextMenu
                    onEdit={() => setDialog("edit")}
                    onDelete={() => setDialog("delete")}
                  />
                </div>

                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      size="sm"
                      className="absolute -bottom-4 left-1/2 z-30 -translate-x-1/2"
                    >
                      <CirclePlus />
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent asChild>
                    <Card
                      className="z-10 -mt-2 flex w-fit flex-col p-2 text-center dark:bg-gray-dark-2"
                      noShadow
                    >
                      {buttons.map((button) => (
                        <Button
                          key={button.label}
                          className="my-1 w-36 justify-start text-xs"
                          onClick={() => {
                            setType(button.type);
                            setDialog("create");
                          }}
                        >
                          <button.icon size={16} />
                          {button.label}
                        </Button>
                      ))}
                    </Card>
                  </PopoverContent>
                </Popover>
                {dialog !== null && (
                  <DialogContent className="z-50">
                    {dialog === "edit" && (
                      <BlockForm
                        block={block}
                        action={dialog}
                        path={blockPath}
                      />
                    )}
                    {dialog === "create" && (
                      <BlockForm
                        block={getPlaceholderBlock(type)}
                        action={dialog}
                        path={blockPath}
                      />
                    )}
                    {dialog === "delete" && (
                      <BlockDeleteForm
                        type={block.type}
                        action={dialog}
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
      </DraggableElement>
    </>
  );
};
