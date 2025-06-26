import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react";

import { AllBlocksType } from "../../../schema/blocks";
import Badge from "~/design/Badge";
import { BlockPathType } from "..";
import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { ParsingError } from "./ParsingError";
import { SelectBlockType } from "../../../BlockEditor/components/SelectType";
import { pathsEqual } from "../../context/utils";
import { twMergeClsx } from "~/util/helpers";
import { useCreateBlock } from "../../context/utils/useCreateBlock";
import { usePageEditor } from "../../context/pageCompilerContext";
import { usePageEditorPanel } from "../../../BlockEditor/EditorPanelProvider";
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
  const { mode, focus, setFocus } = usePageEditor();
  const { setPanel } = usePageEditorPanel();
  const [isHovered, setIsHovered] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const createBlock = useCreateBlock();

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
    setPanel({
      action: "edit",
      block,
      path: blockPath,
    });
    return setFocus(blockPath); // Todo: do we still need focus state separate from panel state?
  };

  const isFocused = focus && pathsEqual(focus, blockPath);

  return (
    <>
      <div
        ref={containerRef}
        className={twMergeClsx(
          mode === "edit" &&
            "relative rounded-md border-2 border-dashed border-gray-4 bg-white p-3 dark:border-gray-dark-4 dark:bg-black",
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
          <Badge className="absolute z-30 -m-6" variant="secondary">
            <b>{block.type}</b>
            {blockPath.join(".")}
          </Badge>
        )}
        {mode === "edit" && isFocused && (
          <div onClick={(event) => event.stopPropagation()}>
            <SelectBlockType
              path={blockPath}
              onSelect={(type) => createBlock(type, blockPath)}
            />
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
