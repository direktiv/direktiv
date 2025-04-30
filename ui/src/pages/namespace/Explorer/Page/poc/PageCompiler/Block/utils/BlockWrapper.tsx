import {
  PropsWithChildren,
  Suspense,
  useEffect,
  useRef,
  useState,
} from "react";

import { AllBlocksType } from "../../../schema/blocks";
import Badge from "~/design/Badge";
import { BlockPath } from "./blockPath";
import { ErrorBoundary } from "react-error-boundary";
import { Loading } from "./Loading";
import { UserError } from "./UserError";
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
    if (mode !== "preview") {
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
    <div
      ref={containerRef}
      className={twMergeClsx(
        mode === "preview" &&
          "rounded-md relative p-3 border-2 border-gray-4 border-dashed dark:border-gray-dark-4 bg-white dark:bg-black",
        isHovered &&
          mode === "preview" &&
          "border-solid bg-gray-2 dark:bg-gray-dark-2"
      )}
      data-block-wrapper
    >
      {mode === "preview" && (
        <Badge
          className="-m-6 absolute z-50"
          variant="secondary"
          style={{
            display: isHovered ? "block" : "none",
          }}
        >
          <b>{block.type}</b> {blockPath}
        </Badge>
      )}
      <Suspense fallback={<Loading />}>
        <ErrorBoundary
          fallbackRender={({ error }) => (
            <UserError title={t("direktivPage.error.queryProvider.apiError")}>
              {error.message}
            </UserError>
          )}
        >
          {children}
        </ErrorBoundary>
      </Suspense>
    </div>
  );
};
