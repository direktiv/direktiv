import { ReactElement, Suspense } from "react";

import { BlockPathType } from "..";
import { Dropzone } from "~/design/DragAndDrop/Dropzone";
import { Loading } from "./Loading";
import { twMergeClsx } from "~/util/helpers";
import { useEnableDropzone } from "./useEnableDropzone";
import { usePageStateContext } from "../../context/pageCompilerContext";

type BlockListProps = {
  horizontal?: boolean;
  children: ReactElement[];
  path: BlockPathType;
};

type WrapperProps = {
  horizontal?: boolean;
  children: ReactElement;
};

type BlockListComponentProps = BlockListProps;

const BlockListWrapper = ({ children, horizontal }: WrapperProps) => (
  <div
    className={twMergeClsx(
      "w-full gap-3",
      horizontal
        ? "grid grid-cols-[repeat(auto-fit,minmax(100px,1fr))]"
        : "flex flex-col"
    )}
  >
    {children}
  </div>
);

const EditorBlockList = ({
  horizontal,
  children,
  path,
}: BlockListComponentProps) => {
  const enableDropZone = useEnableDropzone();

  const newBlockTargetPath = [...path, 0];

  return (
    <BlockListWrapper horizontal={horizontal}>
      <Suspense fallback={<Loading />}>
        {!children.length && (
          <div className="w-full self-center">
            <Dropzone
              enable={enableDropZone}
              payload={{ targetPath: newBlockTargetPath }}
            />
          </div>
        )}
        {children}
      </Suspense>
    </BlockListWrapper>
  );
};

const VisitorBlockList = ({
  horizontal,
  children,
}: BlockListComponentProps) => (
  <BlockListWrapper horizontal={horizontal}>
    <Suspense fallback={<Loading />}>{children}</Suspense>
  </BlockListWrapper>
);

export const BlockList = (props: BlockListComponentProps) => {
  const { mode } = usePageStateContext();

  if (mode === "edit") {
    return <EditorBlockList {...props} />;
  }

  return <VisitorBlockList {...props} />;
};
