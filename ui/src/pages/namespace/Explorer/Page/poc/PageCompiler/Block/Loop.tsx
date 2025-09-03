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
import { BlockList } from "./utils/BlockList";
import { LoopType } from "../../schema/blocks/loop";
import { VariableError } from "../primitives/Variable/Error";
import { usePageStateContext } from "../context/pageCompilerContext";
import { useTranslation } from "react-i18next";
import { useVariableArrayResolver } from "../primitives/Variable/utils/useVariableArrayResolver";

type LoopProps = {
  blockProps: LoopType;
  blockPath: BlockPathType;
};

export const Loop = ({ blockProps, blockPath }: LoopProps) => {
  const { blocks, data, id } = blockProps;
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

  const loopItems =
    mode === "edit" ? variableArray.data.slice(0, 1) : variableArray.data;
  const hiddenItemsCount = variableArray.data.length - 1;
  const showBadge =
    mode === "edit" && blocks.length > 0 && hiddenItemsCount > 0;

  return (
    <BlockList path={blockPath}>
      {loopItems.map((item, variableIndex) => (
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
                <Block key={path.join(".")} block={block} blockPath={path} />
              );
            })}
          </BlockList>
          {showBadge && <LoopItemBadge count={hiddenItemsCount} />}
        </VariableContextProvider>
      ))}
    </BlockList>
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
