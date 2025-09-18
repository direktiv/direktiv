import { Check, Locate, Plus, X } from "lucide-react";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";
import { FC, ReactNode, useMemo, useState } from "react";
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
  label?: string;
  tree: Tree;
  onSubmit: (value: string) => void;
  preventSubmit: (path: string[]) => boolean;
  container?: HTMLDivElement;
  preview: (path: string[]) => ReactNode;
};

export const TreePicker: FC<TreePickerProps> = ({
  label,
  tree,
  container,
  preventSubmit = () => false,
  onSubmit,
  preview,
}) => {
  const { t } = useTranslation();
  const [path, setPath] = useState<string[]>([]);
  const [search, setSearch] = useState("");
  const currentTree = useMemo(() => getSublist(tree, path), [path, tree]);
  const disabled = useMemo(() => preventSubmit(path), [preventSubmit, path]);

  const allowCustomSegment = search.length > 0;
  const formattedPath = path.join(".");

  const addCustomSegment = () => {
    setPath([...path, search]);
    setSearch("");
  };

  return (
    <Popover>
      <div className="flex">
        <PopoverTrigger asChild>
          {label ? (
            <Button
              variant="outline"
              type="button"
              className="dark:border-gray-dark-7"
            >
              {label}
            </Button>
          ) : (
            <Button
              icon
              variant="ghost"
              type="button"
              className="dark:border-gray-dark-7"
            >
              <Locate />
            </Button>
          )}
        </PopoverTrigger>
      </div>
      <PopoverContent className="w-96" align="start" container={container}>
        <Command>
          <CommandInput
            placeholder={t("direktivPage.blockEditor.treePicker.placeholder")}
            value={search}
            onValueChange={setSearch}
            onKeyDown={(event) => {
              if (event.key === "Enter" && allowCustomSegment) {
                addCustomSegment();
              }
            }}
          >
            <Button
              icon
              type="button"
              variant="ghost"
              onClick={addCustomSegment}
              className="-mr-2"
              disabled={!allowCustomSegment}
            >
              <Plus className="text-xs" />
            </Button>
          </CommandInput>
          <CommandList>
            <CommandEmpty>
              <div className="text-sm text-gray-11 dark:text-gray-dark-11">
                {t("direktivPage.blockEditor.treePicker.listEmpty")}
              </div>
            </CommandEmpty>
            {!!currentTree?.length && (
              <CommandGroup
                heading={t("direktivPage.blockEditor.treePicker.valuesHeader")}
              >
                {currentTree.map((key) => (
                  <CommandItem
                    key={key}
                    onSelect={() => setPath([...path, key])}
                  >
                    {key}
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
          <div className="flex items-center p-2">
            <FakeInput
              wrap
              className="mr-2 w-full text-gray-10 dark:text-gray-dark-10"
            >
              {preview(path)}
            </FakeInput>
            <div className="ml-auto flex gap-2">
              <PopoverClose asChild>
                <Button
                  variant="outline"
                  icon
                  type="button"
                  disabled={disabled}
                  onClick={() => {
                    onSubmit(formattedPath);
                    setPath([]);
                  }}
                >
                  <Check />
                </Button>
              </PopoverClose>
              <PopoverClose asChild>
                <Button
                  variant="outline"
                  icon
                  type="button"
                  onClick={() => setPath([])}
                >
                  <X />
                </Button>
              </PopoverClose>
            </div>
          </div>
        </Command>
      </PopoverContent>
    </Popover>
  );
};
