import { Plus, X } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import DatePicker from "./DatePicker";
import { FiltersObj } from "~/api/instances/query/get";
import Options from "./Options";
import RefineTime from "./RefineTime";
import { SelectFieldMenu } from "./SelectFieldMenu";
import TextInput from "./TextInput";
import moment from "moment";
import { useState } from "react";
import { useTranslation } from "react-i18next";

type FiltersProps = {
  filters: FiltersObj;
  onUpdate: (filters: FiltersObj) => void;
};

type MenuAnchor =
  | "main"
  | "AS"
  | "STATUS"
  | "TRIGGER"
  | "AFTER"
  | "BEFORE"
  | "AFTER.time"
  | "BEFORE.time";

const fieldsInMenu: Array<keyof FiltersObj> = [
  "AS",
  "STATUS",
  "TRIGGER",
  "AFTER",
  "BEFORE",
];

const Filters = ({ filters, onUpdate }: FiltersProps) => {
  const { t } = useTranslation();

  // activeMenu controls which popover component is opened (there are
  // separate popovers triggered by the respective buttons)
  const [activeMenu, setActiveMenu] = useState<MenuAnchor | null>(null);

  // selectedField controls which submenu is shown in the main menu
  const [selectedField, setSelectedField] = useState<keyof FiltersObj | null>(
    null
  );

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

  const setFilter = (newFilter: FiltersObj) => {
    const newFilters = { ...filters, ...newFilter };
    onUpdate(newFilters);
    resetMenu();
  };

  const clearFilter = (field: keyof FiltersObj) => {
    const newFilters = { ...filters };
    delete newFilters[field];
    onUpdate(newFilters);
  };

  const currentFilterKeys = Object.keys(filters) as Array<keyof FiltersObj>;

  const hasFilters = !!currentFilterKeys.length;

  const undefinedFilters = fieldsInMenu.filter(
    (field) => !currentFilterKeys.includes(field)
  );

  return (
    <div className="m-2 flex flex-row flex-wrap gap-2">
      {currentFilterKeys.map((field) => {
        // For type safety, one separate return is required below for every type
        // so it is possible to assert filters[field]?.value is defined and TS
        // does not merge the different possible types of filters[field]?.value

        if (field === "AS") {
          return (
            <ButtonBar key={field}>
              <Button variant="outline" asChild>
                <label>
                  {t([`pages.instances.list.filter.field.${field}`])}
                </label>
              </Button>
              <Popover
                open={activeMenu === field}
                onOpenChange={(state) => handleOpenChange(state, field)}
              >
                <PopoverTrigger asChild>
                  <Button variant="outline">{filters[field]?.value}</Button>
                </PopoverTrigger>
                <PopoverContent align="start">
                  {field === "AS" && (
                    <TextInput
                      field={field}
                      setFilter={setFilter}
                      clearFilter={clearFilter}
                      value={filters[field]?.value}
                    />
                  )}
                </PopoverContent>
              </Popover>
              <Button variant="outline" icon>
                <X onClick={() => clearFilter(field)} />
              </Button>
            </ButtonBar>
          );
        }

        if (field === "STATUS" || field === "TRIGGER") {
          return (
            <ButtonBar key={field}>
              <Button variant="outline" asChild>
                <label>
                  {t([`pages.instances.list.filter.field.${field}`])}
                </label>
              </Button>
              <Popover
                open={activeMenu === field}
                onOpenChange={(state) => handleOpenChange(state, field)}
              >
                <PopoverTrigger asChild>
                  <Button variant="outline">{filters[field]?.value}</Button>
                </PopoverTrigger>
                <PopoverContent align="start">
                  {field === "STATUS" && (
                    <Options
                      field={field}
                      value={filters[field]?.value}
                      setFilter={setFilter}
                    />
                  )}
                  {field === "TRIGGER" && (
                    <Options
                      field={field}
                      value={filters[field]?.value}
                      setFilter={setFilter}
                    />
                  )}
                </PopoverContent>
              </Popover>
              <Button variant="outline" icon>
                <X onClick={() => clearFilter(field)} />
              </Button>
            </ButtonBar>
          );
        }

        if (field === "AFTER" || field == "BEFORE") {
          const dateValue = filters[field]?.value;
          if (!dateValue) {
            console.error("Early return: dateValue is not defined");
            return null;
          }
          return (
            <ButtonBar key={field}>
              <Button variant="outline" asChild>
                <label>
                  {t([`pages.instances.list.filter.field.${field}`])}
                </label>
              </Button>
              <Popover
                open={activeMenu === field}
                onOpenChange={(state) => handleOpenChange(state, field)}
              >
                <PopoverTrigger asChild>
                  <Button variant="outline" className="px-2">
                    {moment(filters[field]?.value).format("YYYY-MM-DD")}
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="start">
                  {(field === "AFTER" || field === "BEFORE") && (
                    <DatePicker
                      field={field}
                      date={filters[field]?.value}
                      setFilter={setFilter}
                    />
                  )}
                </PopoverContent>
              </Popover>
              <Popover
                open={activeMenu === `${field}.time`}
                onOpenChange={(state) =>
                  handleOpenChange(state, `${field}.time`)
                }
              >
                <PopoverTrigger asChild>
                  <Button variant="outline" className="px-2">
                    {moment(filters[field]?.value).format("HH:mm:ss")}
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="start" className="w-min">
                  <RefineTime
                    field={field}
                    date={dateValue}
                    setFilter={setFilter}
                  />
                </PopoverContent>
              </Popover>
              <Button variant="outline" icon>
                <X onClick={() => clearFilter(field)} />
              </Button>
            </ButtonBar>
          );
        }
      })}

      {!!undefinedFilters.length && (
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
              <Button variant="outline" onClick={() => toggleMenu("main")}>
                <Plus />
                {t("pages.instances.list.filter.filterButton")}
              </Button>
            )}
          </PopoverTrigger>
          <PopoverContent align="start">
            {(selectedField === null && (
              <SelectFieldMenu
                options={undefinedFilters}
                onSelect={setSelectedField}
              />
            )) ||
              (selectedField === "AS" && (
                <TextInput
                  field={selectedField}
                  setFilter={setFilter}
                  clearFilter={clearFilter}
                  value={filters[selectedField]?.value}
                />
              )) ||
              (selectedField === "STATUS" && (
                <Options
                  field={selectedField}
                  value={filters[selectedField]?.value}
                  setFilter={setFilter}
                />
              )) ||
              (selectedField === "TRIGGER" && (
                <Options
                  field={selectedField}
                  value={filters[selectedField]?.value}
                  setFilter={setFilter}
                />
              )) ||
              ((selectedField === "AFTER" || selectedField === "BEFORE") && (
                <DatePicker
                  field={selectedField}
                  date={filters[selectedField]?.value}
                  setFilter={setFilter}
                />
              ))}
          </PopoverContent>
        </Popover>
      )}
    </div>
  );
};

export default Filters;
