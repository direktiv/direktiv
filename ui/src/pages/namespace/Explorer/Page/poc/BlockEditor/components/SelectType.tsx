import { CirclePlus, Heading1, LucideIcon, Text } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import { AllBlocksType } from "../../schema/blocks";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { Trans } from "react-i18next";
import { t } from "i18next";

type SelectBlockTypeProps = {
  onSelect: (type: AllBlocksType["type"]) => void;
  big?: boolean;
};

const buttons: {
  type: AllBlocksType["type"];
  label: string;
  icon: LucideIcon;
}[] = [
  {
    type: "headline" satisfies AllBlocksType["type"],
    label: t("direktivPage.blockEditor.blockName.headline"),
    icon: Heading1,
  },
  {
    type: "text" satisfies AllBlocksType["type"],
    label: t("direktivPage.blockEditor.blockName.text"),
    icon: Text,
  },
];

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

const Content = ({ onSelect }: Pick<SelectBlockTypeProps, "onSelect">) => (
  <PopoverContent asChild>
    <Card
      className="z-10 -mt-2 flex w-fit flex-col p-2 text-center dark:bg-gray-dark-2"
      noShadow
    >
      {buttons.map((button) => (
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

export const SelectBlockType = ({
  onSelect,
  big = false,
}: SelectBlockTypeProps) => (
  <Popover>
    {big ? <BigTrigger /> : <DefaultTrigger />}
    <Content onSelect={onSelect} />
  </Popover>
);
