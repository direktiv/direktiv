import { Check, Plus } from "lucide-react";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";
import { FC, useMemo, useState } from "react";
import {
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
} from "~/design/Popover";
import { Tree, getSublist } from "./utils";

import Button from "~/design/Button";
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
  const [search, setSearch] = useState("");
  const currentTree = useMemo(() => getSublist(tree, path), [tree, path]);
  const allowSubmit = useMemo(() => path.length, [path]);
  const allowCustomSegment = useMemo(() => search.length > 0, [search]);
  const formattedPath = useMemo(
    () => (path.length ? `{{${path.join(".")}}}` : null),
    [path]
  );

  return (
    <Popover>
      <div className="flex">
        <PopoverTrigger asChild>
          <Button variant="outline" type="button">
            {t("direktivPage.blockEditor.smartInput.variableBtn")}
          </Button>
        </PopoverTrigger>
        <div className="self-center px-3 text-sm text-gray-11">
          {formattedPath}
        </div>
        <PopoverClose>
          <Button
            variant="outline"
            icon
            type="button"
            disabled={!allowSubmit}
            onClick={() => {
              if (formattedPath) {
                onSubmit(formattedPath);
                setPath([]);
              }
            }}
          >
            <Check />
          </Button>
        </PopoverClose>
      </div>
      <PopoverContent align="start" container={container}>
        <Command>
          <CommandInput
            placeholder="select a value"
            value={search}
            onValueChange={setSearch}
          >
            <Button
              icon
              type="button"
              variant="ghost"
              onClick={() => setPath([...path, search])}
              className="pr-0"
              disabled={!allowCustomSegment}
            >
              <Plus size="xs" />
            </Button>
          </CommandInput>
          <CommandList>
            <CommandEmpty>
              <div>{t("direktivPage.blockEditor.smartInput.listEmpty")}</div>
            </CommandEmpty>
            <CommandGroup heading="value">
              {currentTree?.map((key) => (
                <CommandItem key={key} onSelect={() => setPath([...path, key])}>
                  {key}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
        {path.length >= 3 && <div></div>}
      </PopoverContent>
    </Popover>
  );
};
