import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";

import { FilterField } from ".";
import { useTranslation } from "react-i18next";

export const SelectFieldMenu = ({
  options,
  onSelect,
}: {
  onSelect: (value: FilterField) => void;
  options: Array<FilterField>;
}) => {
  const { t } = useTranslation();
  return (
    <Command>
      <CommandInput
        placeholder={t(
          "pages.instances.detail.logs.filter.placeholder.mainMenu"
        )}
      />
      <CommandList>
        <CommandGroup
          heading={t("pages.instances.detail.logs.filter.menuHeading.main")}
        >
          {options.map((option) => (
            <CommandItem key={option} onSelect={() => onSelect(option)}>
              {t(`pages.instances.detail.logs.filter.field.${option}`)}
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  );
};
