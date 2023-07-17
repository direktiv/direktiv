import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";

import { FiltersObj } from "~/api/instances/query/get";
import { useTranslation } from "react-i18next";

export const SelectFieldMenu = ({
  onSelect,
}: {
  onSelect: (value: keyof FiltersObj) => void;
}) => {
  const { t } = useTranslation();
  return (
    <Command>
      <CommandInput
        placeholder={t("pages.instances.list.filter.placeholder.mainMenu")}
      />
      <CommandList>
        <CommandGroup
          heading={t("pages.instances.list.filter.menuHeading.main")}
        >
          <CommandItem tabIndex={1} onSelect={() => onSelect("AS")}>
            {t("pages.instances.list.filter.field.AS")}
          </CommandItem>
          <CommandItem onSelect={() => onSelect("STATUS")}>
            {t("pages.instances.list.filter.field.STATUS")}
          </CommandItem>
          <CommandItem onSelect={() => onSelect("TRIGGER")}>
            {t("pages.instances.list.filter.field.TRIGGER")}
          </CommandItem>
          <CommandItem onSelect={() => onSelect("AFTER")}>
            {t("pages.instances.list.filter.field.AFTER")}
          </CommandItem>
          <CommandItem onSelect={() => onSelect("BEFORE")}>
            {t("pages.instances.list.filter.field.BEFORE")}
          </CommandItem>
        </CommandGroup>
      </CommandList>
    </Command>
  );
};
