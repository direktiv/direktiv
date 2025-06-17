import { CirclePlus, Columns2, Heading1, LucideIcon, Text } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import { AllBlocksType } from "../../schema/blocks";
import { BlockPathType } from "../../PageCompiler/Block";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { Trans } from "react-i18next";
import { t } from "i18next";
import { usePageEditor } from "../../PageCompiler/context/pageCompilerContext";

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
  const { parseAncestors } = usePageEditor();

  const buttons: {
    type: AllBlocksType["type"];
    label: string;
    icon: LucideIcon;
    allow: (path: BlockPathType) => boolean;
  }[] = [
    {
      type: "headline",
      label: t("direktivPage.blockEditor.blockName.headline"),
      icon: Heading1,
      allow: () => true,
    },
    {
      type: "text",
      label: t("direktivPage.blockEditor.blockName.text"),
      icon: Text,
      allow: () => true,
    },
    {
      type: "columns",
      label: t("direktivPage.blockEditor.blockName.columns"),
      icon: Columns2,
      allow: (path: BlockPathType) =>
        !parseAncestors(path, (block) => block.type === "columns"),
    },
  ];

  const filteredButtons = buttons.filter((type) => type.allow(path) === true);

  return (
    <PopoverContent asChild>
      <Card
        className="z-10 -mt-2 flex w-fit flex-col p-2 text-center dark:bg-gray-dark-2"
        noShadow
      >
        {filteredButtons.map((button) => (
          <Button
            variant="outline"
            key={button.label}
            className="my-1 w-36 justify-start text-xs"
            onClick={() => onSelect(button.type)}
          >
            <button.icon size={16} />
            {button.label}
          </Button>
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
