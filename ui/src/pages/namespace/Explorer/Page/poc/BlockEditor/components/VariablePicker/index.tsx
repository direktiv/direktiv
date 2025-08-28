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

import Button from "~/design/Button";
import { Check } from "lucide-react";
import { useTranslation } from "react-i18next";
import z from "zod";

type Tree = {
  [key: string]: Tree | unknown;
};

const TreeSchema: z.ZodType<Tree> = z.lazy(() =>
  z.record(z.union([TreeSchema, z.unknown()]))
);

const isTree = (value: unknown): value is Tree =>
  TreeSchema.safeParse(value).success;

type VariablePickerProps = {
  tree: Tree;
  onSubmit: (value: string) => void;
  container?: HTMLDivElement;
};

const getSubtree = (tree: Tree, path: string[]): Tree =>
  path.reduce<Tree>((current, segment) => {
    const next = current[segment];
    return isTree(next) ? next : current;
  }, tree);

const getSublist = (tree: Tree, path: string[]): string[] => {
  const subtree = getSubtree(tree, path);
  if (typeof subtree === "string" || typeof subtree === "undefined") {
    return [];
  }
  return Object.keys(subtree);
};

export const VariablePicker: FC<VariablePickerProps> = ({
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
