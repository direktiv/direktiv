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
import { ComponentPropsWithoutRef, FC, useState } from "react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "../../../../../../design/Popover";

import Button from "../../../../../../design/Button";
import { TrimedRevisionSchemaType } from "../../../../../../api/tree/schema";
import clsx from "clsx";
import { useTranslation } from "react-i18next";

const RevisionSelector: FC<
  ComponentPropsWithoutRef<typeof Button> & {
    tags: TrimedRevisionSchemaType[];
    revisions: TrimedRevisionSchemaType[];
    isLoading?: boolean;
    onSelect?: (revision: string) => void;
  }
> = ({ tags, revisions, isLoading, onSelect, ...props }) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [value, setValue] = useState("");
  const revisionsWithoutTags = revisions.filter(
    (rev) => !tags.some((t) => t.name === rev.name)
  );

  const tagsAndRevisions = [...revisions]; // revisions have tags include

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
          {value
            ? tagsAndRevisions.find((rev) => rev.name === value)?.name
            : t(
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
                  onSelect={(currentValue) => {
                    setValue(currentValue === value ? "" : currentValue);
                    onSelect?.(currentValue === value ? "" : currentValue);
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
                  onSelect={(currentValue) => {
                    setValue(currentValue === value ? "" : currentValue);
                    onSelect?.(currentValue === value ? "" : currentValue);
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
