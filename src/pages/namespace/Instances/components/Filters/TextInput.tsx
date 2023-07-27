import { Command, CommandGroup, CommandList } from "~/design/Command";

import { FiltersObj } from "~/api/instances/query/get";
import Input from "~/design/Input";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const TextInput = ({
  field,
  value,
  setFilter,
  clearFilter,
}: {
  field: keyof FiltersObj;
  value?: string;
  setFilter: (filter: FiltersObj) => void;
  clearFilter: (field: keyof FiltersObj) => void;
}) => {
  const [inputValue, setInputValue] = useState<string>(value || "");
  const { t } = useTranslation();

  const handleKeyDown = (event: { key: string }) => {
    // Currently API only supports CONTAINS on filter fields with text inputs
    const type = "CONTAINS";

    if (event.key === "Enter" && inputValue) {
      setFilter({
        [field]: { value: inputValue, type },
      });
    }
    if (event.key === "Enter" && !inputValue) {
      clearFilter(field);
    }
  };

  return (
    <Command>
      <CommandList>
        <CommandGroup heading={t("pages.instances.list.filter.menuHeading.AS")}>
          <Input
            autoFocus
            placeholder={t("pages.instances.list.filter.placeholder.AS")}
            value={inputValue}
            onChange={(event) => setInputValue(event.target.value)}
            onKeyDown={handleKeyDown}
          />
        </CommandGroup>
      </CommandList>
    </Command>
  );
};

export default TextInput;
