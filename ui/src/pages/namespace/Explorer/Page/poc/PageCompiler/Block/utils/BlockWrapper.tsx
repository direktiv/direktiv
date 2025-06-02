import { CirclePlus, Edit } from "lucide-react";
import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react";

import { AllBlocksType } from "../../../schema/blocks";
import Badge from "~/design/Badge";
import { BlockForm } from "../../../BlockEditor";
import { BlockPathType } from "..";
import Button from "~/design/Button";
import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { pathsEqual } from "../../context/utils";
import { twMergeClsx } from "~/util/helpers";
import { usePageEditor } from "../../context/pageCompilerContext";
import { useTranslation } from "react-i18next";

type BlockWrapperProps = PropsWithChildren<{
  blockPath: BlockPathType;
  block: AllBlocksType;
}>;

type DialogState = "create" | "edit" | null;

export const BlockWrapper = ({
  block,
  blockPath,
  children,
}: BlockWrapperProps) => {
  const { t } = useTranslation();
  const { mode, focus, setFocus } = usePageEditor();
  const [dialog, setDialog] = useState<DialogState>(null);
  const [isHovered, setIsHovered] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

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
            "relative rounded-md p-3 border-2 border-gray-4 border-dashed dark:border-gray-dark-4 bg-white dark:bg-black",
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
          <Badge className="-m-6 absolute z-30" variant="secondary">
            <b>{block.type}</b>
            {blockPath.join(".")}
          </Badge>
        )}
        {mode === "edit" && isFocused && (
          <div onClick={(event) => event.stopPropagation()}>
            <Dialog open={!!dialog} onOpenChange={handleOnOpenChange}>
              <DialogTrigger asChild onClick={() => setDialog("edit")}>
                <Button variant="ghost" className="absolute right-1 top-1 z-30">
                  <Edit />
                </Button>
              </DialogTrigger>
              <DialogTrigger className="float-right" asChild>
                <Button
                  size="sm"
                  className="absolute -bottom-4 z-30 right-1/2"
                  onClick={() => setDialog("create")}
                >
                  <CirclePlus />
                </Button>
              </DialogTrigger>
              {dialog !== null && (
                <DialogContent className="z-50">
                  {dialog === "edit" && (
                    <BlockForm
                      block={block}
                      action={dialog}
                      path={blockPath}
                      close={() => setDialog(null)}
                    />
                  )}
                  {dialog === "create" && (
                    <BlockForm
                      block={{ type: "text", content: "dummy block" }}
                      action={dialog}
                      path={blockPath}
                      close={() => setDialog(null)}
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
