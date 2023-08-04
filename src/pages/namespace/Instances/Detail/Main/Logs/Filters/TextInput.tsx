import { Command, CommandGroup, CommandList } from "~/design/Command";

import { FilterField } from ".";
import Input from "~/design/Input";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const TextInput = ({
  field,
  value,
  setFilter,
  clearFilter,
}: {
  field: FilterField;
  value?: string;
  setFilter: (field: FilterField, value: string) => void;
  clearFilter: (field: FilterField) => void;
}) => {
  const [inputValue, setInputValue] = useState<string>(value || "");
  const { t } = useTranslation();

  const handleKeyDown = (event: { key: string }) => {
    if (event.key === "Enter" && inputValue) {
      setFilter(field, inputValue);
    }
    if (event.key === "Enter" && !inputValue) {
      clearFilter(field);
    }
  };

  return (
    <Command>
      <CommandList>
        <CommandGroup
          heading={t(`pages.instances.detail.logs.filter.menuHeading.${field}`)}
        >
          <Input
            autoFocus
            placeholder={t(
              `pages.instances.detail.logs.filter.placeholder.${field}`
            )}
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
