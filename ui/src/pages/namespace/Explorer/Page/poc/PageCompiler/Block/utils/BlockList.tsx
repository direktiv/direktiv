import { ReactElement, Suspense, useMemo } from "react";

import { BlockPathType } from "..";
import { Dropzone } from "~/design/DragAndDrop/Dropzone";
import { Loading } from "./Loading";
import { twMergeClsx } from "~/util/helpers";
import { usePageStateContext } from "../../context/pageCompilerContext";
import { useValidateDropzone } from "./useValidateDropzone";

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
      "w-full",
      horizontal
        ? "grid grid-cols-[repeat(auto-fit,minmax(100px,1fr))] gap-6"
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
  const validateDropzone = useValidateDropzone();

  const dropzonePayload = useMemo(() => ({ targetPath: [...path, 0] }), [path]);

  return (
    <BlockListWrapper horizontal={horizontal}>
      <Suspense fallback={<Loading />}>
        {!children.length && (
          <div className="flex h-full min-h-[25px] flex-col justify-center">
            <Dropzone validate={validateDropzone} payload={dropzonePayload} />
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
