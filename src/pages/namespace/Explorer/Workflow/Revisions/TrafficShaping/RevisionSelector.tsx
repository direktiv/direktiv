import { ChevronDown, Circle, GitMerge, Tags } from "lucide-react";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandStaticItem,
} from "../../../../../../design/Command";
import { ComponentPropsWithoutRef, FC, useEffect, useState } from "react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "../../../../../../design/Popover";

import Button from "../../../../../../design/Button";
import { TrimmedRevisionSchemaType } from "../../../../../../api/tree/schema";
import clsx from "clsx";
import { useTranslation } from "react-i18next";

// use props of our button but overwrite some onSelect and defaultValue
type ButtonProps = Omit<
  ComponentPropsWithoutRef<typeof Button>,
  // remove the native onSelect and defaultValue props,
  // asChild also must be unset to allow a dynamic loading prop (asChild does not support loading prop)
  "onSelect" | "defaultValue" | "asChild"
> & {
  onSelect?: (revision: string) => void;
  defaultValue?: string;
};

type RevisionSelectorProps = ButtonProps & {
  tags: TrimmedRevisionSchemaType[];
  revisions: TrimmedRevisionSchemaType[];
  isLoading?: boolean;
};

const RevisionSelector: FC<RevisionSelectorProps> = ({
  tags,
  revisions,
  isLoading,
  onSelect,
  defaultValue,
  ...props
}) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [value, setValue] = useState(defaultValue ?? "");
  const [isTagSelected, setIsTagSelected] = useState(false);

  const revisionsWithoutTags = revisions.filter(
    (rev) => !tags.some((t) => t.name === rev.name)
  );

  // when defaultValue is changing, synch it with component state
  useEffect(() => {
    setValue(defaultValue ?? "");
    setIsTagSelected(tags.some((t) => t.name === defaultValue));
  }, [defaultValue, tags]);

  const buttonLabel =
    value !== ""
      ? revisions
          .find((rev) => rev.name === value)
          ?.name.slice(0, isTagSelected ? undefined : 8)
      : undefined;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          loading={isLoading}
          {...props}
        >
          {buttonLabel ??
            t(
              "pages.explorer.tree.workflow.revisions.trafficShaping.revisionSelector.placeholder"
            )}
          <ChevronDown />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[360px]">
        <Command>
          <CommandList className="max-h-[378px]">
            <CommandInput
              placeholder={t(
                "pages.explorer.tree.workflow.revisions.trafficShaping.revisionSelector.searchPlaceholder"
              )}
            />
            <CommandStaticItem className="text-sm font-semibold text-gray-9 dark:text-gray-dark-9">
              <Tags className="mr-2 h-auto w-4" />
              {t(
                "pages.explorer.tree.workflow.revisions.trafficShaping.revisionSelector.tags"
              )}
            </CommandStaticItem>
            <CommandGroup>
              {tags.map((tag) => (
                <CommandItem
                  value={tag.name}
                  key={tag.name}
                  data-testid={tag.name}
                  onSelect={(currentValue) => {
                    setValue(currentValue === value ? "" : currentValue);
                    onSelect?.(currentValue === value ? "" : currentValue);
                    setIsTagSelected(true);
                    setOpen(false);
                  }}
                >
                  <Circle
                    className={clsx(
                      "mr-2 h-2 w-2 fill-current",
                      value === tag.name ? "opacity-100" : "opacity-0"
                    )}
                  />
                  {tag.name}
                </CommandItem>
              ))}
            </CommandGroup>
            <CommandStaticItem className="text-sm font-semibold text-gray-9 dark:text-gray-dark-9">
              <GitMerge className="mr-2 h-auto w-4" />
              {t(
                "pages.explorer.tree.workflow.revisions.trafficShaping.revisionSelector.revisions"
              )}
            </CommandStaticItem>
            <CommandGroup>
              {revisionsWithoutTags.map((revision) => (
                <CommandItem
                  value={revision.name}
                  key={revision.name}
                  data-testid={revision.name}
                  onSelect={(currentValue) => {
                    setValue(currentValue === value ? "" : currentValue);
                    onSelect?.(currentValue === value ? "" : currentValue);
                    setIsTagSelected(false);
                    setOpen(false);
                  }}
                >
                  <Circle
                    className={clsx(
                      "mr-2 h-2 w-2 fill-current",
                      value === revision.name ? "opacity-100" : "opacity-0"
                    )}
                  />
                  {revision.name}
                </CommandItem>
              ))}
            </CommandGroup>
            <CommandEmpty>
              {t(
                "pages.explorer.tree.workflow.revisions.trafficShaping.revisionSelector.notFound"
              )}
            </CommandEmpty>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
};

export default RevisionSelector;
