import { Command, CommandGroup, CommandList } from "~/design/Command";

import { ArrowRight } from "lucide-react";
import Button from "~/design/Button";
import { FilterField } from ".";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
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

  const applyFilter = () => {
    if (inputValue) {
      setFilter(field, inputValue);
    } else {
      clearFilter(field);
    }
  };

  const handleKeyDown = (event: { key: string }) => {
    if (event.key === "Enter") {
      applyFilter();
    }
  };

  return (
    <Command>
      <CommandList>
        <CommandGroup
          heading={t(`pages.instances.detail.logs.filter.menuHeading.${field}`)}
        >
          <InputWithButton>
            <Input
              autoFocus
              placeholder={t(
                `pages.instances.detail.logs.filter.placeholder.${field}`
              )}
              value={inputValue}
              onChange={(event) => setInputValue(event.target.value)}
              onKeyDown={handleKeyDown}
            />
            <Button icon variant="ghost" onClick={() => applyFilter()}>
              <ArrowRight />
            </Button>
          </InputWithButton>
        </CommandGroup>
      </CommandList>
    </Command>
  );
};

export default TextInput;
