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
      console.log("🚀 clear");
      clearFilter(field);
    }
  };

  return (
    <Command>
      <CommandList>
        {/* TODO: make this i18n dynamic */}
        <CommandGroup heading={field}>
          <Input
            autoFocus
            // TODO: make this i18n dynamic
            placeholder={field}
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
