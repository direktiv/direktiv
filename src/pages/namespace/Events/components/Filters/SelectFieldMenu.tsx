import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";

import { FiltersObj } from "~/api/events/query/get";
import { useTranslation } from "react-i18next";

export const SelectFieldMenu = ({
  options,
  onSelect,
}: {
  onSelect: (value: keyof FiltersObj) => void;
  options: Array<keyof FiltersObj>;
}) => {
  const { t } = useTranslation();
  return (
    <Command>
      <CommandInput
        placeholder={t("pages.events.list.filter.placeholder.mainMenu")}
      />
      <CommandList>
        <CommandGroup heading={t("pages.events.list.filter.menuHeading.main")}>
          {options.map((option) => (
            <CommandItem key={option} onSelect={() => onSelect(option)}>
              {t(`pages.events.list.filter.field.${option}`)}
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  );
};
