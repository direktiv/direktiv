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
import { BlockPath } from "..";
import Button from "~/design/Button";
import { Edit } from "lucide-react";
import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { twMergeClsx } from "~/util/helpers";
import { useMode } from "../../context/pageCompilerContext";
import { useTranslation } from "react-i18next";

type BlockWrapperProps = PropsWithChildren<{
  blockPath: BlockPath;
  block: AllBlocksType;
}>;

export const BlockWrapper = ({
  children,
  block,
  blockPath,
}: BlockWrapperProps) => {
  const { t } = useTranslation();
  const mode = useMode();
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

  return (
    <>
      <div
        ref={containerRef}
        className={twMergeClsx(
          mode === "inspect" &&
            "rounded-md relative p-3 border-2 border-gray-4 border-dashed dark:border-gray-dark-4 bg-white dark:bg-black",
          isHovered &&
            mode === "inspect" &&
            "border-solid bg-gray-2 dark:bg-gray-dark-2"
        )}
        data-block-wrapper
      >
        {mode === "inspect" && (
          <Dialog>
            <Badge
              className="-m-6 absolute z-50"
              variant="secondary"
              style={{
                display: isHovered ? "block" : "none",
              }}
            >
              <b>{block.type}</b> {blockPath.join(".")}
            </Badge>
            <DialogTrigger className="float-right" asChild>
              <Button
                variant="ghost"
                style={{ display: isHovered ? "block" : "none" }}
              >
                <Edit />
              </Button>
            </DialogTrigger>
            <DialogContent>
              <BlockForm path={blockPath}></BlockForm>
            </DialogContent>
          </Dialog>
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
      {mode === "inspect" && (
        <div className="rounded-md border border-gray-7 bg-gray-3 p-1 text-xs text-gray-8">
          Drop Area
        </div>
      )}
    </>
  );
};
