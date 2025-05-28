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
import { CreateBlockForm } from "../../../BlockEditor/create";
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

export const BlockWrapper = ({
  block,
  blockPath,
  children,
}: BlockWrapperProps) => {
  const { t } = useTranslation();
  const { mode, focus, setFocus, addBlock } = usePageEditor();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [isHovered, setIsHovered] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (mode !== "inspect") {
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
    if (mode !== "inspect") {
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
          mode === "inspect" &&
            "relative rounded-md p-3 border-2 border-gray-4 border-dashed dark:border-gray-dark-4 bg-white dark:bg-black",
          isHovered &&
            mode === "inspect" &&
            "border-solid bg-gray-2 dark:bg-gray-dark-2",
          isFocused &&
            mode === "inspect" &&
            "border-solid border-gray-8 dark:border-gray-10"
        )}
        data-block-wrapper
        onClick={handleClickBlock}
      >
        {mode === "inspect" && (isHovered || isFocused) && (
          <Badge className="-m-6 absolute z-30" variant="secondary">
            <b>{block.type}</b>
            {blockPath.join(".")}
          </Badge>
        )}
        {mode === "inspect" && isFocused && (
          <>
            <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
              <DialogTrigger
                asChild
                onClick={(event) => {
                  event.stopPropagation();
                  setDialogOpen(true);
                }}
              >
                <Button variant="ghost" className="absolute right-1 top-1 z-30">
                  <Edit />
                </Button>
              </DialogTrigger>
              <DialogContent className="z-50">
                <BlockForm
                  path={blockPath}
                  close={() => setDialogOpen(false)}
                ></BlockForm>
              </DialogContent>
            </Dialog>
            <Dialog>
              <DialogTrigger className="float-right" asChild>
                <Button size="sm" className="absolute -bottom-4 z-30 right-1/2">
                  <CirclePlus />
                </Button>
              </DialogTrigger>
              <DialogContent>
                <CreateBlockForm
                  setSelectedBlock={(newBlock) => {
                    addBlock(
                      blockPath,
                      {
                        ...newBlock,
                      },
                      true
                    );
                  }}
                  path={blockPath}
                />
              </DialogContent>
            </Dialog>
          </>
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
