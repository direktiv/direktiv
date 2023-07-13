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

  // TODO: This is currently hard coded for field "AS", but
  // should be usable for other fields
  const handleKeyDown = (event: { key: string }) => {
    if (event.key === "Enter" && inputValue) {
      setFilter({
        AS: { value: inputValue, type: "CONTAINS" },
      });
    }
    if (event.key === "Enter" && !inputValue) {
      clearFilter("AS");
    }
  };

  return (
    <>
      {field === "AS" && (
        <Command>
          <CommandList>
            <CommandGroup heading="Filter by name">
              <Input
                autoFocus
                placeholder="filename.yaml"
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
          <CommandInput autoFocus placeholder="Type a command or search..." />
          <CommandList>
            <CommandGroup heading="Filter by state">
              <CommandItem
                value="pending"
                onSelect={() =>
                  setFilter({
                    STATUS: { value: "pending", type: "MATCH" },
                  })
                }
              >
                Running
              </CommandItem>
              <CommandItem
                value="complete"
                onSelect={() =>
                  setFilter({
                    STATUS: { value: "complete", type: "MATCH" },
                  })
                }
              >
                Complete
              </CommandItem>
              <CommandItem
                value="cancelled"
                onSelect={() =>
                  setFilter({
                    STATUS: { value: "cancelled", type: "MATCH" },
                  })
                }
              >
                Cancelled
              </CommandItem>
              <CommandItem
                value="failed"
                onSelect={() =>
                  setFilter({
                    STATUS: { value: "failed", type: "MATCH" },
                  })
                }
              >
                Failed
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      )}
      {field === "TRIGGER" && (
        <Command value={value}>
          <CommandInput autoFocus placeholder="Type a command or search..." />
          <CommandList>
            <CommandGroup heading="Filter by invoker">
              <CommandItem
                value="api"
                onSelect={() =>
                  setFilter({
                    TRIGGER: { value: "api", type: "MATCH" },
                  })
                }
              >
                API
              </CommandItem>
              <CommandItem
                value="cloudevent"
                onSelect={() =>
                  setFilter({
                    TRIGGER: { value: "cloudevent", type: "MATCH" },
                  })
                }
              >
                Cloud event
              </CommandItem>
              <CommandItem
                value="instance"
                onSelect={() =>
                  setFilter({
                    TRIGGER: { value: "instance", type: "MATCH" },
                  })
                }
              >
                Instance
              </CommandItem>
              <CommandItem
                value="cron"
                onSelect={() =>
                  setFilter({
                    TRIGGER: { value: "cron", type: "MATCH" },
                  })
                }
              >
                Cron
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      )}
      {field === "AFTER" && (
        <Command>
          <CommandList className="max-h-[460px]">
            <CommandGroup heading="Filter created after">
              <Datepicker />
            </CommandGroup>
          </CommandList>
        </Command>
      )}
      {field === "BEFORE" && (
        <Command>
          <CommandList className="max-h-[460px]">
            <CommandGroup heading="Filter created before">
              <Datepicker />
            </CommandGroup>
          </CommandList>
        </Command>
      )}
    </>
  );
};

export default FieldSubMenu;
