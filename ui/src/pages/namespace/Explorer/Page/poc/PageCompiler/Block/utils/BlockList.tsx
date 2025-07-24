import { ReactElement, Suspense } from "react";
import { pathIsDescendant, pathsEqual } from "../../context/utils";

import { BlockPathType } from "..";
import { DragPayloadSchemaType } from "~/design/DragAndDrop/schema";
import { Dropzone } from "~/design/DragAndDrop/Dropzone";
import { Loading } from "./Loading";
import { twMergeClsx } from "~/util/helpers";
import { useBlockTypes } from "../../context/utils/useBlockTypes";
import { usePageEditorPanel } from "../../../BlockEditor/EditorPanelProvider";
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
  const { panel } = usePageEditorPanel();
  const { getAllowedTypes } = useBlockTypes();

  const newBlockTargetPath = [...path, 0];

  const enableDropZone = (payload: DragPayloadSchemaType | null) => {
    if (panel?.dialog && !pathIsDescendant(path, panel.dialog)) {
      return false;
    }

    const allowedTypes = getAllowedTypes(newBlockTargetPath);
    if (!allowedTypes.some((config) => config.type === payload?.blockType)) {
      return false;
    }

    if (payload?.type === "move") {
      // don't show a dropzone for neighboring blocks
      if (
        pathsEqual(payload.originPath, newBlockTargetPath) ||
        pathsEqual(payload.originPath, path)
      ) {
        return false;
      }
    }
    return true;
  };

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
