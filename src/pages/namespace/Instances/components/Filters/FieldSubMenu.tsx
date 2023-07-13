import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";
import { FilterField, FiltersObj } from ".";

import { Datepicker } from "~/design/Datepicker";
import Input from "~/design/Input";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const FieldSubMenu = ({
  field,
  value,
  setFilter,
  clearFilter,
}: {
  field: FilterField;
  value?: string;
  setFilter: (filter: FiltersObj) => void;
  clearFilter: (field: FilterField) => void;
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
    <>
      {field === "AS" && (
        <Command>
          <CommandList>
            <CommandGroup
              heading={t("pages.instances.list.filter.menuHeading.AS")}
            >
              <Input
                autoFocus
                placeholder={t("pages.instances.list.filter.placeholder.AS")}
                value={inputValue}
                onChange={(event) => setInputValue(event.target.value)}
                onKeyUp={handleKeyDown}
              />
            </CommandGroup>
          </CommandList>
        </Command>
      )}
      {field === "STATUS" && (
        <Command value={value}>
          <CommandInput
            autoFocus
            placeholder={t("pages.instances.list.filter.placeholder.STATUS")}
          />
          <CommandList>
            <CommandGroup
              heading={t("pages.instances.list.filter.menuHeading.STATUS")}
            >
              <CommandItem
                value="pending"
                onSelect={() =>
                  setFilter({
                    STATUS: { value: "pending", type: "MATCH" },
                  })
                }
              >
                {t("pages.instances.list.filter.status.pending")}
              </CommandItem>
              <CommandItem
                value="complete"
                onSelect={() =>
                  setFilter({
                    STATUS: { value: "complete", type: "MATCH" },
                  })
                }
              >
                {t("pages.instances.list.filter.status.complete")}
              </CommandItem>
              <CommandItem
                value="cancelled"
                onSelect={() =>
                  setFilter({
                    STATUS: { value: "cancelled", type: "MATCH" },
                  })
                }
              >
                {t("pages.instances.list.filter.status.cancelled")}
              </CommandItem>
              <CommandItem
                value="failed"
                onSelect={() =>
                  setFilter({
                    STATUS: { value: "failed", type: "MATCH" },
                  })
                }
              >
                {t("pages.instances.list.filter.status.failed")}
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      )}
      {field === "TRIGGER" && (
        <Command value={value}>
          <CommandInput
            autoFocus
            placeholder={t("pages.instances.list.filter.placeholder.TRIGGER")}
          />
          <CommandList>
            <CommandGroup
              heading={t("pages.instances.list.filter.menuHeading.TRIGGER")}
            >
              <CommandItem
                value="api"
                onSelect={() =>
                  setFilter({
                    TRIGGER: { value: "api", type: "MATCH" },
                  })
                }
              >
                {t("pages.instances.list.filter.trigger.api")}
              </CommandItem>
              <CommandItem
                value="cloudevent"
                onSelect={() =>
                  setFilter({
                    TRIGGER: { value: "cloudevent", type: "MATCH" },
                  })
                }
              >
                {t("pages.instances.list.filter.trigger.cloudevent")}
              </CommandItem>
              <CommandItem
                value="instance"
                onSelect={() =>
                  setFilter({
                    TRIGGER: { value: "instance", type: "MATCH" },
                  })
                }
              >
                {t("pages.instances.list.filter.trigger.instance")}
              </CommandItem>
              <CommandItem
                value="cron"
                onSelect={() =>
                  setFilter({
                    TRIGGER: { value: "cron", type: "MATCH" },
                  })
                }
              >
                {t("pages.instances.list.filter.trigger.cron")}
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      )}
      {field === "AFTER" && (
        <Command>
          <CommandList className="max-h-[460px]">
            <CommandGroup
              heading={t("pages.instances.list.filter.menuHeading.AFTER")}
            >
              <Datepicker />
            </CommandGroup>
          </CommandList>
        </Command>
      )}
      {field === "BEFORE" && (
        <Command>
          <CommandList className="max-h-[460px]">
            <CommandGroup
              heading={t("pages.instances.list.filter.menuHeading.BEFORE")}
            >
              <Datepicker />
            </CommandGroup>
          </CommandList>
        </Command>
      )}
    </>
  );
};

export default FieldSubMenu;
