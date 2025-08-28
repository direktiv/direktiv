import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";
import { FC, useState } from "react";
import {
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
} from "~/design/Popover";
import { Tree, getSublist } from "./utils";

import Button from "~/design/Button";
import { Check } from "lucide-react";
import { useTranslation } from "react-i18next";

type TreePickerProps = {
  tree: Tree;
  onSubmit: (value: string) => void;
  container?: HTMLDivElement;
};

export const TreePicker: FC<TreePickerProps> = ({
  tree,
  container,
  onSubmit,
}) => {
  const { t } = useTranslation();

  const [path, setPath] = useState<string[]>([]);

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" type="button">
          {t("direktivPage.blockEditor.smartInput.variableBtn")}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="start" container={container}>
        {!path.length && (
          <Command>
            <CommandInput placeholder="select variable namespace" />
            <CommandList>
              <CommandGroup heading="namespace">
                {Object.entries(tree).map(([namespace]) => (
                  <CommandItem
                    key={namespace}
                    onSelect={() => setPath([namespace])}
                  >
                    {namespace}
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        )}
        {!!path.length && (
          <Command>
            <CommandInput placeholder="select block id" />
            <CommandList>
              <CommandGroup heading="block scope">
                {getSublist(tree, path).map((id) => (
                  <CommandItem key={id} onSelect={() => setPath([...path, id])}>
                    {id}
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        )}
        {path.length === 2 && (
          <div>
            <PopoverClose>
              <Button
                variant="outline"
                icon
                type="button"
                onClick={() => {
                  onSubmit(` {{${path.join(".")}}}`);
                  setPath([]);
                }}
              >
                <Check />
              </Button>
            </PopoverClose>
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
};
