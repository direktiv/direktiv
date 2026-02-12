import { Block, BlockPathType } from ".";
import { Eye, EyeOff } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";
import {
  VariableContextProvider,
  useVariablesContext,
} from "../primitives/Variable/VariableContext";

import Badge from "~/design/Badge";
import { BlockList } from "page-blocklist";
import { LoopType } from "../../schema/blocks/loop";
import { Pagination } from "~/components/Pagination";
import PaginationProvider from "~/components/PaginationProvider";
import { StopPropagation } from "~/components/StopPropagation";
import { VariableError } from "../primitives/Variable/Error";
import { usePageStateContext } from "../context/pageCompilerContext";
import { useTranslation } from "react-i18next";
import { useVariableArrayResolver } from "../primitives/Variable/utils/useVariableArrayResolver";

type LoopProps = {
  blockProps: LoopType;
  blockPath: BlockPathType;
};

export const Loop = ({ blockProps, blockPath }: LoopProps) => {
  const { blocks, data, id, pageSize } = blockProps;
  const { t } = useTranslation();
  const { mode } = usePageStateContext();
  const resolveVariableArray = useVariableArrayResolver();
  const parentVariables = useVariablesContext();

  const variableArray = resolveVariableArray(data);

  if (parentVariables.loop?.[id]) {
    throw new Error(t("direktivPage.error.duplicateId", { id }));
  }

  if (!variableArray.success) {
    return (
      <VariableError value={data} errorCode={variableArray.error}>
        {t(`direktivPage.error.templateString.${variableArray.error}`, {
          variable: data,
        })}{" "}
        ({variableArray.error})
      </VariableError>
    );
  }

  const hiddenItemsCount = pageSize - 1;
  const showBadge =
    mode === "edit" && blocks.length > 0 && hiddenItemsCount > 0;

  return (
    <PaginationProvider items={variableArray.data} pageSize={pageSize}>
      {({ currentItems, goToPage, goToFirstPage, currentPage, totalPages }) => {
        if (currentPage > totalPages) {
          goToFirstPage();
        }

        const itemsToRender =
          mode === "edit" ? currentItems.slice(0, 1) : currentItems;

        return (
          <>
            <BlockList path={blockPath}>
              {itemsToRender.map((item, variableIndex) => (
                <VariableContextProvider
                  key={variableIndex}
                  variables={{
                    ...parentVariables,
                    loop: {
                      ...parentVariables.loop,
                      [id]: item,
                    },
                  }}
                >
                  <BlockList path={blockPath}>
                    {blocks.map((block, blockIndex) => {
                      const path = [...blockPath, blockIndex];
                      return (
                        <Block
                          key={path.join(".")}
                          block={block}
                          blockPath={path}
                        />
                      );
                    })}
                  </BlockList>
                  {showBadge && <LoopItemBadge count={hiddenItemsCount} />}
                </VariableContextProvider>
              ))}
            </BlockList>
            <div className="flex items-center justify-end gap-2">
              <StopPropagation>
                <div className="mt-2 pb-4">
                  <Pagination
                    totalPages={totalPages}
                    value={currentPage}
                    onChange={goToPage}
                  />
                </div>
              </StopPropagation>
            </div>
          </>
        );
      }}
    </PaginationProvider>
  );
};

const LoopItemBadge = ({ count }: { count: number }) => {
  const { t } = useTranslation();
  return (
    <div className="absolute -bottom-1 flex h-[2px] w-full flex-col items-center justify-center">
      <Badge className="bg-gray-4 text-black dark:bg-gray-dark-4 dark:text-white">
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger className="flex">
              <EyeOff className="mr-2" size={16} />
              {t(`direktivPage.page.blocks.loop.hiddenItems`, {
                count,
              })}
            </TooltipTrigger>
            <TooltipContent>
              {t(`direktivPage.page.blocks.loop.infoHiddenItems_first`)}
              <Eye className="mx-1 inline" size={16} />
              {t(`direktivPage.page.blocks.loop.infoHiddenItems_second`)}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </Badge>
    </div>
  );
};
