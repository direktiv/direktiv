import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";
import {
  FiltersObj,
  StatusValue,
  TriggerValue,
  statusValues,
  triggerValues,
} from "~/api/instances/query/get";

import { useTranslation } from "react-i18next";

const optionMenus = {
  STATUS: statusValues,
  TRIGGER: triggerValues,
};

type OptionsProps = {
  setFilter: (filter: FiltersObj) => void;
} & (
  | {
      value?: TriggerValue;
      field: "TRIGGER";
    }
  | {
      value?: StatusValue;
      field: "STATUS";
    }
);

const Options = ({ value, field, setFilter }: OptionsProps) => {
  const { t } = useTranslation();
  return (
    <Command value={value}>
      <CommandInput
        autoFocus
        placeholder={t("pages.instances.list.filter.placeholder.STATUS")}
      />
      <CommandList>
        <CommandGroup
          heading={t("pages.instances.list.filter.menuHeading.STATUS")}
        >
          {optionMenus[field].map((option) => (
            <CommandItem
              key={option}
              value={option}
              onSelect={() =>
                setFilter({
                  [field]: {
                    value: option,
                    // TODO: Move this decision to the API layer?
                    type: field === "TRIGGER" ? "CONTAINS" : "MATCH",
                  },
                })
              }
            >
              {option}
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  );
};

export default Options;
