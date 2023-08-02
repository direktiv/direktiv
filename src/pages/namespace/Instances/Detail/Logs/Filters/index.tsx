import { Plus, X } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";
import { useActions, useFilters } from "../../state/instanceContext";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import Input from "~/design/Input";
import { SelectFieldMenu } from "./SelectFieldMenu";
import TextInput from "./TextInput";
import { useState } from "react";

const filterFields = ["workflowName", "stateName"] as const;

export type FilterField = (typeof filterFields)[number];
type MenuAnchor = "main" | FilterField;

const Filters = () => {
  const {
    updateFilterStateName,
    updateFilterWorkflow,
    resetFilterStateName,
    resetFilterWorkflow,
  } = useActions();
  const filters = useFilters();

  // activeMenu controls which popover component is opened (there are
  // separate popovers triggered by the respective buttons)
  const [activeMenu, setActiveMenu] = useState<MenuAnchor | null>(null);

  // selectedField controls which submenu is shown in the main menu
  const [selectedField, setSelectedField] = useState<FilterField | null>(null);

  const handleOpenChange = (isOpening: boolean, menu: MenuAnchor) => {
    if (!isOpening) {
      setSelectedField(null);
    }
    toggleMenu(menu);
  };

  const toggleMenu = (value: MenuAnchor) => {
    if (activeMenu === value) {
      return setActiveMenu(null);
    }
    setActiveMenu(value);
  };

  const resetMenu = () => {
    setActiveMenu(null);
    setSelectedField(null);
  };

  const currentFilterKeys = filterFields.filter(
    (items) => filters?.QUERY?.[items]
  );

  const hasFilters = false; // TODO: implement
  const undefinedFilters = filterFields.filter((x) => x); // TODO: implement filtering

  return (
    <ButtonBar>
      {currentFilterKeys.map((field) => (
        <ButtonBar key={field}>
          <Button variant="outline" size="sm">
            {field}
          </Button>
          <Popover
            open={activeMenu === field}
            onOpenChange={(state) => handleOpenChange(state, field)}
          >
            <PopoverTrigger asChild>
              <Button variant="outline" size="sm">
                {filters?.["QUERY"]?.[field]}
              </Button>
            </PopoverTrigger>
            <PopoverContent align="start">
              <TextInput
                field={field}
                // setFilter={setFilter}
                // clearFilter={clearFilter}
                // value={filters[field]?.value}

                setFilter={(filter, value) => {
                  if (filter === "workflowName") {
                    updateFilterWorkflow(value);
                  }
                  if (filter === "stateName") {
                    updateFilterStateName(value);
                  }
                }}
                clearFilter={(filter) => {
                  if (filter === "workflowName") {
                    resetFilterWorkflow();
                  }
                  if (filter === "stateName") {
                    resetFilterStateName();
                  }
                }}
                value={filters?.QUERY?.[field]}
              />
            </PopoverContent>
          </Popover>
          <Button variant="outline" size="sm" icon>
            <X
              onClick={() => {
                if (field === "workflowName") {
                  resetFilterWorkflow();
                }
                if (field === "stateName") {
                  resetFilterStateName();
                }
              }}
            />
          </Button>
        </ButtonBar>
      ))}
      <Popover
        open={activeMenu === "main"}
        onOpenChange={(state) => handleOpenChange(state, "main")}
      >
        <PopoverTrigger asChild>
          {hasFilters ? (
            <Button variant="outline" icon onClick={() => toggleMenu("main")}>
              <Plus />
            </Button>
          ) : (
            <Button
              variant="outline"
              size="sm"
              onClick={() => toggleMenu("main")}
            >
              <Plus />
              filter
            </Button>
          )}
        </PopoverTrigger>
        <PopoverContent align="start">
          {selectedField === null ? (
            <SelectFieldMenu
              options={undefinedFilters}
              onSelect={setSelectedField}
            />
          ) : (
            <TextInput
              field={selectedField}
              setFilter={(filter, value) => {
                if (filter === "workflowName") {
                  updateFilterWorkflow(value);
                }
                if (filter === "stateName") {
                  updateFilterStateName(value);
                }
              }}
              clearFilter={(filter) => {
                if (filter === "workflowName") {
                  resetFilterWorkflow();
                }
                if (filter === "stateName") {
                  resetFilterStateName();
                }
              }}
              value={filters?.QUERY?.[selectedField]}
            />
          )}
        </PopoverContent>
      </Popover>
    </ButtonBar>
  );
};

export default Filters;
