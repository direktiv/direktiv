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
import { FakeInput } from "~/design/FakeInput";
import { useTranslation } from "react-i18next";

type TreePickerProps = {
  tree: Tree;
  onSubmit: (value: string) => void;
  container?: HTMLDivElement;
  placeholders?: string[];
};

export const TreePicker: FC<TreePickerProps> = ({
  tree,
  container,
  placeholders = [],
  onSubmit,
}) => {
  const { t } = useTranslation();
  const [path, setPath] = useState<string[]>([]);
  const [search, setSearch] = useState("");
  const currentTree = useMemo(() => getSublist(tree, path), [tree, path]);
  const allowSubmit = useMemo(() => path.length, [path]);
  const allowCustomSegment = useMemo(() => search.length > 0, [search]);
  const formattedPath = useMemo(() => `{{${path.join(".")}}}`, [path]);
  const previewPath = useMemo(
    () =>
      placeholders.map((placeholder, index) => (
        <>
          {path[index] ? (
            <span key={index} className="text-gray-12">
              {path[index]}
            </span>
          ) : (
            <span key={index} className="text-gray-11">
              {placeholder}
            </span>
          )}
          {index < placeholders.length - 1 && (
            <span className="text-gray-11">.</span>
          )}
        </>
      )),
    [path, placeholders]
  );

  return (
    <Popover>
      <div className="flex">
        <PopoverTrigger asChild>
          <Button variant="outline" type="button">
            {t("direktivPage.blockEditor.smartInput.variableBtn")}
          </Button>
        </PopoverTrigger>
      </div>
      <PopoverContent align="start" container={container}>
        <Command>
          <CommandInput
            placeholder={t("direktivPage.blockEditor.smartInput.placeholder")}
            value={search}
            onValueChange={setSearch}
          >
            <Button
              icon
              type="button"
              variant="ghost"
              onClick={() => setPath([...path, search])}
              className="-mr-2"
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
          <div className="flex items-center p-2">
            <FakeInput className="mr-2 w-full">{previewPath}</FakeInput>
            <PopoverClose className="ml-auto">
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
        </Command>
      </PopoverContent>
    </Popover>
  );
};
