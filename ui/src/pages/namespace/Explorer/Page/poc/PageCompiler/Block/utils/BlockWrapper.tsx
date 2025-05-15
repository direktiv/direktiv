import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react";
import { useBlock, useMode } from "../../context/pageCompilerContext";

import { AllBlocksType } from "../../../schema/blocks";
import Badge from "~/design/Badge";
import { BlockPath } from "./blockPath";
import Button from "~/design/Button";
import { Edit } from "lucide-react";
import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { twMergeClsx } from "~/util/helpers";

// const cloneBlocks = (blocks: AllBlocksType[]): AllBlocksType[] =>
//   structuredClone(blocks);

type BlockWrapperProps = PropsWithChildren<{
  blockPath: BlockPath;
  block: AllBlocksType;
}>;

// type Block = AllBlocksType;
// type List = AllBlocksType[];
// type BlockOrList = Block | List;

// const isParentBlock = (
//   block: AllBlocksType
// ): block is z.infer<typeof ParentBlockUnion> =>
//   ParentBlockUnion.safeParse(block).success;

// const getBlock = (list: BlockOrList, path: BlockPath): BlockOrList => {
//   const result = path.reduce<BlockOrList>((acc, index) => {
//     let next;

//     if (Array.isArray(acc)) {
//       next = acc[index];
//     } else if (isParentBlock(acc)) {
//       next = acc.blocks[index];
//     }

//     if (next) {
//       return next;
//     }

//     throw Error(`index ${index} not found in ${JSON.stringify(acc)}`);
//   }, list);
//   return result;
// };

const BlockForm = ({ path }: { path: BlockPath }) => {
  // const page = usePage();
  const block = useBlock(path);
  return (
    <div>
      Block form for {path} from {JSON.stringify(block)}
    </div>
  );
};

export const BlockWrapper = ({
  children,
  block,
  blockPath,
}: BlockWrapperProps) => {
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
            <DialogTrigger className="float-right">
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
              <ParsingError title="There was an error fetching data from the API">
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
