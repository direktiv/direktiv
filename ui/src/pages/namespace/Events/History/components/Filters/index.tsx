import { Plus, X } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import DatePicker from "~/components/Filters/DatePicker";
import { FiltersSchemaType } from "~/api/events/schema/filters";
import RefineTime from "~/components/Filters/RefineTime";
import { SelectFieldMenu } from "~/components/Filters/SelectFieldMenu";
import TextInput from "~/components/Filters/TextInput";
import moment from "moment";
import { useState } from "react";
import { useTranslation } from "react-i18next";

type FiltersProps = {
  filters: FiltersSchemaType;
  onUpdate: (filters: FiltersSchemaType) => void;
};

type MenuAnchor =
  | "main"
  | "typeContains"
  | "eventContains"
  | "receivedAfter"
  | "receivedBefore"
  | "receivedAfter.time"
  | "receivedBefore.time";

const fieldsInMenu = [
  "typeContains",
  "eventContains",
  "receivedAfter",
  "receivedBefore",
] as const;

const Filters = ({ filters, onUpdate }: FiltersProps) => {
  const { t } = useTranslation();

  // activeMenu controls which popover component is opened (there are
  // separate popovers triggered by the respective buttons)
  const [activeMenu, setActiveMenu] = useState<MenuAnchor | null>(null);

  // selectedField controls which submenu is shown in the main menu
  const [selectedField, setSelectedField] = useState<
    keyof FiltersSchemaType | null
  >(null);

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

  const setFilter = (newFilter: FiltersSchemaType) => {
    const newFilters = { ...filters, ...newFilter };
    onUpdate(newFilters);
    resetMenu();
  };

  const clearFilter = (field: keyof FiltersSchemaType) => {
    const newFilters = { ...filters };
    delete newFilters[field];
    onUpdate(newFilters);
  };

  const currentFilterKeys = Object.keys(filters) as Array<
    keyof FiltersSchemaType
  >;

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

        if (field === "typeContains" || field === "eventContains") {
          return (
            <ButtonBar key={field}>
              <Button variant="outline" asChild>
                <label>
                  {t([`pages.events.history.filter.field.${field}`])}
                </label>
              </Button>
              <Popover
                open={activeMenu === field}
                onOpenChange={(state) => handleOpenChange(state, field)}
              >
                <PopoverTrigger asChild>
                  <Button variant="outline">{filters[field]}</Button>
                </PopoverTrigger>
                <PopoverContent align="start">
                  <TextInput
                    value={filters[field]}
                    onSubmit={(value) => {
                      if (value) {
                        setFilter({
                          [field]: value,
                        });
                      } else {
                        clearFilter(field);
                      }
                    }}
                    heading={t(
                      `pages.events.history.filter.menuHeading.${field}`
                    )}
                    placeholder={t(
                      `pages.events.history.filter.placeholder.${field}`
                    )}
                  />
                </PopoverContent>
              </Popover>
              <Button
                variant="outline"
                icon
                data-testid={`clear-filter-${field}`}
              >
                <X onClick={() => clearFilter(field)} />
              </Button>
            </ButtonBar>
          );
        }

        if (field === "receivedAfter" || field == "receivedBefore") {
          const dateValue = filters[field];
          if (!dateValue) {
            console.error("Early return: dateValue is not defined");
            return null;
          }
          return (
            <ButtonBar key={field}>
              <Button variant="outline" asChild>
                <label>
                  {t([`pages.events.history.filter.field.${field}`])}
                </label>
              </Button>
              <Popover
                open={activeMenu === field}
                onOpenChange={(state) => handleOpenChange(state, field)}
              >
                <PopoverTrigger asChild>
                  <Button variant="outline" className="px-2">
                    {moment(filters[field]).format("YYYY-MM-DD")}
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="start">
                  {(field === "receivedAfter" ||
                    field === "receivedBefore") && (
                    <DatePicker
                      date={filters[field]}
                      heading={t(`components.filters.menuHeading.${field}`)}
                      onChange={(value) =>
                        setFilter({
                          [field]: value,
                        })
                      }
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
                    {moment(filters[field]).format("HH:mm:ss")}
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="start" className="w-min">
                  <RefineTime
                    date={dateValue}
                    onChange={(newDate) => {
                      setFilter({
                        [field]: newDate,
                      });
                    }}
                  />
                </PopoverContent>
              </Popover>
              <Button
                variant="outline"
                icon
                data-testid={`clear-filter-${field}`}
              >
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
              <Button
                data-testid="add-filter"
                variant="outline"
                icon
                onClick={() => toggleMenu("main")}
              >
                <Plus />
              </Button>
            ) : (
              <Button
                data-testid="add-filter"
                variant="outline"
                onClick={() => toggleMenu("main")}
              >
                <Plus />
                {t("pages.events.history.filter.filterButton")}
              </Button>
            )}
          </PopoverTrigger>
          <PopoverContent align="start">
            {(selectedField === null && (
              <SelectFieldMenu<keyof FiltersSchemaType>
                options={undefinedFilters.map((option) => ({
                  value: option,
                  label: t(`pages.events.history.filter.field.${option}`),
                }))}
                onSelect={(value) => setSelectedField(value)}
                heading={t("pages.events.history.filter.menuHeading.main")}
                placeholder={t(
                  "pages.events.history.filter.placeholder.mainMenu"
                )}
              />
            )) ||
              ((selectedField === "typeContains" ||
                selectedField === "eventContains") && (
                <TextInput
                  value={filters[selectedField]}
                  onSubmit={(value) => {
                    if (value) {
                      setFilter({
                        [selectedField]: value,
                      });
                    } else {
                      clearFilter(selectedField);
                    }
                  }}
                  heading={t(
                    `pages.events.history.filter.menuHeading.${selectedField}`
                  )}
                  placeholder={t(
                    `pages.events.history.filter.placeholder.${selectedField}`
                  )}
                />
              )) ||
              ((selectedField === "receivedAfter" ||
                selectedField === "receivedBefore") && (
                <DatePicker
                  heading={t(`components.filters.menuHeading.${selectedField}`)}
                  date={filters[selectedField]}
                  onChange={(value) =>
                    setFilter({
                      [selectedField]: value,
                    })
                  }
                />
              ))}
          </PopoverContent>
        </Popover>
      )}
    </div>
  );
};

export default Filters;
