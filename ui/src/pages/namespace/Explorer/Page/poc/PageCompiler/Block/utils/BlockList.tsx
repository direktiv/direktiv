import {
  PageCompilerMode,
  usePageEditor,
} from "../../context/pageCompilerContext";
import { ReactElement, Suspense } from "react";

import { BlockPlaceholder } from "../../../BlockEditor/components/Placeholder";
import { Loading } from "./Loading";
import { twMergeClsx } from "~/util/helpers";

type BlockListProps = {
  horizontal?: boolean;
  children: ReactElement[];
};

type BlockListComponentProps = BlockListProps & { mode?: PageCompilerMode };

const BlockListComponent = ({
  horizontal,
  children,
  mode,
}: BlockListComponentProps) => (
  <div
    className={twMergeClsx(
      "gap-3",
      horizontal
        ? "grid grid-cols-[repeat(auto-fit,minmax(100px,1fr))]"
        : "flex flex-col"
    )}
  >
    <Suspense fallback={<Loading />}>
      {mode === "edit" && !children.length && <BlockPlaceholder />}
      {children}
    </Suspense>
  </div>
);

export const BlockList = (args: BlockListProps) => {
  const { mode } = usePageEditor();

  return <BlockListComponent {...{ ...args, mode }} />;
};
