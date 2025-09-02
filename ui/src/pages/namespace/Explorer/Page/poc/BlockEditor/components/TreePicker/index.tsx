import { Check, Plus, X } from "lucide-react";
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
  minDepth?: number;
  container?: HTMLDivElement;
  placeholders?: string[];
};

export const TreePicker: FC<TreePickerProps> = ({
  tree,
  container,
  placeholders = [],
  minDepth = 0,
  onSubmit,
}) => {
  const { t } = useTranslation();
  const [path, setPath] = useState<string[]>([]);
  const [search, setSearch] = useState("");
  const currentTree = useMemo(() => getSublist(tree, path), [path, tree]);
  const allowSubmit = useMemo(() => path.length >= minDepth, [path, minDepth]);
  const allowCustomSegment = useMemo(() => search.length > 0, [search]);
  const formattedPath = useMemo(() => `{{${path.join(".")}}}`, [path]);

  const previewPath = useMemo(() => {
    const previewLength = Math.max(placeholders.length, path.length);
    return Array.from({ length: previewLength }).map((_, index) => (
      <>
        {path[index] ? (
          <span key={index} className="text-gray-12">
            {path[index]}
          </span>
        ) : (
          <span key={index} className="text-gray-10">
            {placeholders[index]}
          </span>
        )}
        {index < previewLength - 1 && <span className="text-gray-10">.</span>}
      </>
    ));
  }, [path, placeholders]);

  const addCustomSegment = () => {
    setPath([...path, search]);
    setSearch("");
  };

  return (
    <Popover>
      <div className="flex">
        <PopoverTrigger asChild>
          <Button variant="outline" type="button">
            {t("direktivPage.blockEditor.smartInput.variableBtn")}
          </Button>
        </PopoverTrigger>
      </div>
      <PopoverContent className="w-96" align="start" container={container}>
        <Command>
          <CommandInput
            placeholder={t("direktivPage.blockEditor.smartInput.placeholder")}
            value={search}
            onValueChange={setSearch}
            onKeyUp={(event) => {
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
              <Plus size="xs" />
            </Button>
          </CommandInput>
          <CommandList>
            <CommandEmpty>
              <div className="text-sm text-gray-11">
                {t("direktivPage.blockEditor.smartInput.listEmpty")}
              </div>
            </CommandEmpty>
            {currentTree && (
              <CommandGroup
                heading={t("direktivPage.blockEditor.smartInput.valuesHeader")}
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
            <FakeInput wrap className="mr-2 w-full text-gray-10">
              {"{{"}
              {previewPath}
              {"}}"}
            </FakeInput>
            <PopoverClose className="ml-auto flex gap-2">
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
        </Command>
      </PopoverContent>
    </Popover>
  );
};
