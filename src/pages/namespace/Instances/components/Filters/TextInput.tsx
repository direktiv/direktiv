import { Command, CommandGroup, CommandList } from "~/design/Command";

import { ArrowRight } from "lucide-react";
import Button from "~/design/Button";
import { FiltersObj } from "~/api/instances/query/get";
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
  field: keyof FiltersObj;
  value?: string;
  setFilter: (filter: FiltersObj) => void;
  clearFilter: (field: keyof FiltersObj) => void;
}) => {
  const [inputValue, setInputValue] = useState<string>(value || "");
  const { t } = useTranslation();

  // Currently API only supports CONTAINS on filter fields with text inputs
  const type = "CONTAINS";

  const applyFilter = () => {
    if (inputValue) {
      setFilter({
        [field]: { value: inputValue, type },
      });
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
        <CommandGroup heading={t("pages.instances.list.filter.menuHeading.AS")}>
          <InputWithButton>
            <Input
              autoFocus
              placeholder={t("pages.instances.list.filter.placeholder.AS")}
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
