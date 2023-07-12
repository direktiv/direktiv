import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";
import { Plus, X } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Datepicker } from "~/design/Datepicker";
import Input from "~/design/Input";
import { useState } from "react";

export type FilterField = "AS" | "STATUS" | "TRIGGER" | "AFTER" | "BEFORE";

type FilterItem = {
  type: "MATCH" | "CONTAINS";
  value: string;
};

export type FiltersObj = {
  [key in FilterField]?: FilterItem;
};

const ParamSelect = ({
  onSelect,
}: {
  onSelect: (value: FilterField) => void;
}) => (
  <Command>
    <CommandInput placeholder="Type a command or search..." />
    <CommandList>
      <CommandGroup heading="Select filter">
        <CommandItem tabIndex={1} onSelect={() => onSelect("AS")}>
          by name
        </CommandItem>
        <CommandItem onSelect={() => onSelect("STATUS")}>by state</CommandItem>
        <CommandItem onSelect={() => onSelect("TRIGGER")}>
          by invoker
        </CommandItem>
        <CommandItem onSelect={() => onSelect("AFTER")}>
          created after
        </CommandItem>
        <CommandItem onSelect={() => onSelect("BEFORE")}>
          created before
        </CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
);

const Filters = ({ onUpdate }: { onUpdate: (filters: FiltersObj) => void }) => {
  const [selectedField, setSelectedField] = useState<FilterField | undefined>();
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const [filters, setFilters] = useState<FiltersObj>({});
  const [inputValue, setInputValue] = useState<string>("");

  const handleOpenChange = (isOpening: boolean) => {
    if (!isOpening) {
      setSelectedField(undefined);
    }
    setIsOpen(isOpening);
  };

  const resetMenu = () => {
    setIsOpen(false);
    setSelectedField(undefined);
  };

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

  const setFilter = (filterObj: FiltersObj) => {
    const newFilters = { ...filters, ...filterObj };
    setFilters(newFilters);
    resetMenu();
    onUpdate(newFilters);
  };

  const clearFilter = (field: FilterField) => {
    const newFilters = { ...filters };
    delete newFilters[field];
    setFilters(newFilters);
    onUpdate(newFilters);
  };

  const hasFilters = !!Object.keys(filters).length;

  const definedFields = Object.keys(filters) as Array<FilterField>;

  return (
    <div className="m-2 flex flex-row gap-2">
      {definedFields.map((field) => (
        <ButtonBar key={field}>
          <Button variant="outline">{field}</Button>
          <Button variant="outline">{filters[field]?.value}</Button>
          <Button variant="outline" icon>
            <X onClick={() => clearFilter(field)} />
          </Button>
        </ButtonBar>
      ))}

      <Popover open={isOpen} onOpenChange={handleOpenChange}>
        <PopoverTrigger asChild>
          {hasFilters ? (
            <Button variant="outline" icon>
              <Plus />
            </Button>
          ) : (
            <Button variant="outline">
              <Plus />
              Filter
            </Button>
          )}
        </PopoverTrigger>
        <PopoverContent align="start">
          {selectedField === undefined && (
            <ParamSelect onSelect={setSelectedField} />
          )}
          {selectedField === "AS" && (
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
          {selectedField === "STATUS" && (
            <Command>
              <CommandInput
                autoFocus
                placeholder="Type a command or search..."
              />
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
          {selectedField === "TRIGGER" && (
            <Command>
              <CommandInput
                autoFocus
                placeholder="Type a command or search..."
              />
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
          {selectedField === "AFTER" && (
            <Command>
              <CommandList className="max-h-[460px]">
                <CommandGroup heading="Filter created after">
                  <Datepicker />
                </CommandGroup>
              </CommandList>
            </Command>
          )}
          {selectedField === "BEFORE" && (
            <Command>
              <CommandList className="max-h-[460px]">
                <CommandGroup heading="Filter created before">
                  <Datepicker />
                </CommandGroup>
              </CommandList>
            </Command>
          )}
        </PopoverContent>
      </Popover>
    </div>
  );
};

export default Filters;
