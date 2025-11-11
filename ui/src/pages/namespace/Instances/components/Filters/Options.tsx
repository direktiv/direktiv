import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";
import {
  FiltersObj,
  InvokerValue,
  StatusValue,
  invokerValues,
  statusValues,
} from "~/api/instances/query/utils";

import { useTranslation } from "react-i18next";

const optionMenus = {
  status: statusValues,
  invoker: invokerValues,
};

type OptionsProps = {
  setFilter: (filter: FiltersObj) => void;
} & (
  | {
      value?: InvokerValue;
      field: "invoker";
    }
  | {
      value?: StatusValue;
      field: "status";
    }
);

const Options = ({ value, field, setFilter }: OptionsProps) => {
  const { t } = useTranslation();
  return (
    <Command value={value}>
      <CommandInput
        autoFocus
        placeholder={t("pages.instances.list.filter.placeholder.status")}
      />
      <CommandList>
        <CommandGroup
          heading={t("pages.instances.list.filter.menuHeading.status")}
        >
          {optionMenus[field].map((option) => (
            <CommandItem
              key={option}
              value={option}
              onSelect={() =>
                setFilter({
                  [field]: {
                    value: option,
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
