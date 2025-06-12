import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import { AllBlocksType } from "../../schema/blocks";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { CirclePlus } from "lucide-react";
import { buttons } from "../../PageCompiler/context/utils";

type SelectBlockTypeProps = { onSelect: (type: AllBlocksType["type"]) => void };

export const SelectBlockType = ({ onSelect }: SelectBlockTypeProps) => (
  <Popover>
    <PopoverTrigger asChild>
      <Button
        size="sm"
        className="absolute -bottom-4 left-1/2 z-30 -translate-x-1/2"
      >
        <CirclePlus />
      </Button>
    </PopoverTrigger>
    <PopoverContent asChild>
      <Card
        className="z-10 -mt-2 flex w-fit flex-col p-2 text-center dark:bg-gray-dark-2"
        noShadow
      >
        {buttons.map((button) => (
          <Button
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
  </Popover>
);
