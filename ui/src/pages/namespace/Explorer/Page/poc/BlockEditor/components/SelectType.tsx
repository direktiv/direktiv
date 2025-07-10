import {
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
} from "~/design/Popover";

import { AllBlocksType } from "../../schema/blocks";
import { BlockPathType } from "../../PageCompiler/Block";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { CirclePlus } from "lucide-react";
import { Trans } from "react-i18next";
import { useBlockTypes } from "../../PageCompiler/context/utils/useBlockTypes";
import { usePage } from "../../PageCompiler/context/pageCompilerContext";

type SelectBlockTypeProps = {
  onSelect: (type: AllBlocksType["type"]) => void;
  big?: boolean;
  path: BlockPathType;
};

const BigTrigger = () => (
  <PopoverTrigger asChild>
    <Button icon variant="outline" className="">
      <CirclePlus />
      <Trans i18nKey="direktivPage.blockEditor.generic.addBlock" />
    </Button>
  </PopoverTrigger>
);

const DefaultTrigger = () => (
  <PopoverTrigger asChild>
    <Button
      variant="primary"
      size="sm"
      className="absolute -bottom-4 left-1/2 z-30 -translate-x-1/2"
    >
      <CirclePlus />
    </Button>
  </PopoverTrigger>
);

const List = ({ onSelect, path }: Omit<SelectBlockTypeProps, "label">) => {
  const page = usePage();
  const blockTypes = useBlockTypes();
  const allowedBlockTypes = blockTypes.filter((type) => type.allow(page, path));
  return (
    <PopoverContent asChild>
      <Card
        className="z-40 -mt-2 flex w-fit flex-col p-2 text-center dark:bg-gray-dark-2"
        noShadow
      >
        {allowedBlockTypes.map((type) => (
          <PopoverClose key={type.label} asChild>
            <Button
              variant="outline"
              className="my-1 w-40 justify-start text-xs"
              onClick={() => onSelect(type.type)}
            >
              <type.icon size={16} />
              {type.label}
            </Button>
          </PopoverClose>
        ))}
      </Card>
    </PopoverContent>
  );
};

export const SelectBlockType = ({
  onSelect,
  big = false,
  path,
}: SelectBlockTypeProps) => (
  <Popover>
    {big ? <BigTrigger /> : <DefaultTrigger />}
    <List onSelect={onSelect} path={path} />
  </Popover>
);
