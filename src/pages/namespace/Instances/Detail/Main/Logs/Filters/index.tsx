import { Plus, X } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";
import { useActions, useFilters } from "../../../store/instanceContext";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { FiltersObj } from "~/api/logs/query/get";
import { SelectFieldMenu } from "./SelectFieldMenu";
import TextInput from "./TextInput";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const availableFilters: Array<keyof FiltersObj> = ["workflowName", "stateName"];
export type FilterField = (typeof availableFilters)[number];
type MenuAnchor = "main" | FilterField;

const Filters = () => {
  const {
    updateFilterStateName,
    updateFilterWorkflow,
    resetFilterStateName,
    resetFilterWorkflow,
  } = useActions();
  const filters = useFilters();
  const { t } = useTranslation();

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

  const setFilter = (field: FilterField, value: string) => {
    if (field === "workflowName") {
      updateFilterWorkflow(value);
    }
    if (field === "stateName") {
      updateFilterStateName(value);
    }
    resetMenu();
  };

  const clearFilter = (field: FilterField) => {
    if (field === "workflowName") {
      resetFilterWorkflow();
    }
    if (field === "stateName") {
      resetFilterStateName();
    }
  };

  const currentFilterKeys = availableFilters.filter(
    (items) => filters?.[items]
  );

  const hasFilters = !!currentFilterKeys.length;
  const undefinedFilters = availableFilters.filter(
    (x) => !currentFilterKeys.includes(x)
  );

  return (
    <>
      {currentFilterKeys.map((field) => (
        <ButtonBar key={field}>
          <Button variant="outline" size="sm" asChild>
            <label>
              {t(`pages.instances.detail.logs.filter.field.${field}`)}
            </label>
          </Button>
          <Popover
            open={activeMenu === field}
            onOpenChange={(state) => handleOpenChange(state, field)}
          >
            <PopoverTrigger asChild>
              <Button variant="outline" size="sm">
                {filters?.[field]}
              </Button>
            </PopoverTrigger>
            <PopoverContent align="start">
              <TextInput
                field={field}
                setFilter={setFilter}
                clearFilter={clearFilter}
                value={filters?.[field]}
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

      {!!undefinedFilters.length && (
        <Popover
          open={activeMenu === "main"}
          onOpenChange={(state) => handleOpenChange(state, "main")}
        >
          <PopoverTrigger asChild>
            {hasFilters ? (
              <Button
                variant="outline"
                size="sm"
                icon
                onClick={() => toggleMenu("main")}
              >
                <Plus />
              </Button>
            ) : (
              <Button
                variant="outline"
                size="sm"
                onClick={() => toggleMenu("main")}
              >
                <Plus />
                {t("pages.instances.detail.logs.filter.filterButton")}
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
                setFilter={setFilter}
                clearFilter={clearFilter}
                value={filters?.[selectedField]}
              />
            )}
          </PopoverContent>
        </Popover>
      )}
    </>
  );
};

export default Filters;
