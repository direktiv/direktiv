import { Block, BlockPathType } from ".";
import { FC, PropsWithChildren } from "react";
import {
  VariableContextProvider,
  useVariablesContext,
} from "../primitives/Variable/VariableContext";

import Badge from "~/design/Badge";
import { BlockList } from "./utils/BlockList";
import { EyeOff } from "lucide-react";
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

  if (parentVariables.loop[id]) {
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

  if (mode === "edit") {
    const loopSize = variableArray.data.length - 1;
    const firstLoopItem = variableArray.data[0];

    return (
      <VariableContextProvider
        variables={{
          ...parentVariables,
          loop: {
            ...parentVariables.loop,
            [id]: firstLoopItem,
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
        <LoopItemBadge>
          {t(`direktivPage.page.blocks.loop.hiddenItems`, {
            number: loopSize,
          })}
        </LoopItemBadge>
      </VariableContextProvider>
    );
  }

  return (
    <BlockList path={blockPath}>
      {variableArray.data.map((item, variableIndex) => (
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
        </VariableContextProvider>
      ))}
    </BlockList>
  );
};

const LoopItemBadge: FC<PropsWithChildren> = ({ children }) => (
  <div className="absolute flex flex-col justify-center items-center mt-1 h-[2px] w-full">
        <Badge className="bg-gray-4 text-black dark:bg-gray-dark-4 dark:text-white">
          <EyeOff className="mr-2" size={16} />
          <span>{children}</span>
        </Badge>
  </div>
);
