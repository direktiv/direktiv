import { Plus, Settings } from "lucide-react";
import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react";
import {
  useMode,
  usePage,
  useSetPage,
} from "../../context/pageCompilerContext";

import { AllBlocksType } from "../../../schema/blocks";
import Badge from "~/design/Badge";
import { BlockPath } from "./blockPath";
import Button from "~/design/Button";
import { ErrorBoundary } from "react-error-boundary";
import { HeadlineType } from "../../../schema/blocks/headline";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { twMergeClsx } from "~/util/helpers";

type BlockWrapperProps = PropsWithChildren<{
  blockPath: BlockPath;
  block: AllBlocksType;
}>;

export const BlockWrapper = ({
  children,
  block,
  blockPath,
}: BlockWrapperProps) => {
  const mode = useMode();
  const page = usePage();
  const setPage = useSetPage();

  const [isHovered, setIsHovered] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const blockPathNumber = Number(blockPath.slice(7));

  const exampleBlock: HeadlineType = {
    type: "headline",
    label: "example",
    size: "h2",
  };

  const addSelectedBlockToPage = (block: HeadlineType, index: number) => {
    const newPage = {
      ...page,
      blocks: [
        ...page.blocks.slice(0, index),
        block,
        ...page.blocks.slice(index),
      ],
    };

    setPage(newPage);
    return newPage;
  };

  const addBlockToPage = () => {
    addSelectedBlockToPage(exampleBlock, blockPathNumber);
  };

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
      <Button
        variant="outline"
        className="w-fit"
        onClick={() => addBlockToPage()}
      >
        <Plus className="size-4 mr-2" /> Add Element
      </Button>
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
          <>
            <Badge
              className="-m-6 absolute z-50"
              variant="secondary"
              style={{
                display: isHovered ? "block" : "none",
              }}
            >
              <b>{block.type}</b> {blockPath}
            </Badge>
            <Button
              variant="ghost"
              style={{ display: isHovered ? "block" : "none" }}
              className="float-right"
            >
              <Settings />
            </Button>
          </>
        )}
        <Suspense fallback={<Loading />}>
          <ErrorBoundary
            fallbackRender={({ error }) => (
              <ParsingError title="There was an error fetching data from the API">
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
