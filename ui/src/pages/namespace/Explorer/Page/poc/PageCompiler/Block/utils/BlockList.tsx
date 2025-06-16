import {
  PageCompilerMode,
  usePageEditor,
} from "../../context/pageCompilerContext";
import { ReactElement, Suspense } from "react";

import { Loading } from "./Loading";
import { SelectBlockType } from "../../../BlockEditor/components/SelectType";
import { getPlaceholderBlock } from "../../context/utils";
import { twMergeClsx } from "~/util/helpers";
import { useBlockDialog } from "../../../BlockEditor/BlockDialogProvider";

type BlockListProps = {
  horizontal?: boolean;
  children: ReactElement[];
};

type BlockListComponentProps = BlockListProps & { mode?: PageCompilerMode };

const BlockListComponent = ({
  horizontal,
  children,
  mode,
}: BlockListComponentProps) => {
  const { setDialog } = useBlockDialog();

  return (
    <div
      className={twMergeClsx(
        "gap-3",
        horizontal
          ? "grid grid-cols-[repeat(auto-fit,minmax(100px,1fr))]"
          : "flex flex-col"
      )}
    >
      <Suspense fallback={<Loading />}>
        {mode === "edit" && !children.length && (
          <div className="self-center">
            <SelectBlockType
              big
              onSelect={(type) =>
                setDialog({
                  action: "create",
                  block: getPlaceholderBlock(type),
                  path: [0],
                })
              }
            />
          </div>
        )}
        {children}
      </Suspense>
    </div>
  );
};

export const BlockList = (args: BlockListProps) => {
  const { mode } = usePageEditor();

  return <BlockListComponent {...{ ...args, mode }} />;
};
