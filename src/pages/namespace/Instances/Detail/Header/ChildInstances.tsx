import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";
import { ConditionalWrapper, twMergeClsx } from "~/util/helpers";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Badge from "~/design/Badge";
import Button from "~/design/Button";
import { RefreshCcw } from "lucide-react";
import { pages } from "~/util/router/pages";
import { statusToBadgeVariant } from "../../utils";
import { t } from "i18next";
import { useInstanceId } from "../store/instanceContext";
import { useInstances } from "~/api/instances/query/get";
import { useNamespace } from "~/util/store/namespace";
import { useState } from "react";

const maxChildInstancesToShow = 50;

const ChildInstances = () => {
  const instanceId = useInstanceId();
  const namespace = useNamespace();
  const [popoverOpen, setPopoverOpen] = useState(false);
  const { data, refetch, isFetching } = useInstances({
    limit: maxChildInstancesToShow + 1,
    offset: 0,
    filters: {
      TRIGGER: {
        type: "MATCH",
        // @ts-expect-error TODO: [DIR-724] currently the type at FiltersObj is not correct at this poinr, we need to update the type and make the filter component of the instance list deal with it
        value: `instance:${instanceId}`,
      },
    },
  });

  if (!namespace) return null;

  const onInstanceSelect = (instance: string) => {
    // useNavigate(); is not working for some reason
    location.href = pages.instances.createHref({ namespace, instance });
  };

  const childCount = data?.instances.results.length ?? 0;

  const needsPopover = childCount > 0;
  const moreInstances = childCount > maxChildInstancesToShow;

  // need to return an element with a height to avoid layout shifts because of css grid
  if (!data) return <>&nbsp;</>;

  return (
    <div className="text-sm">
      <div className="text-gray-10 dark:text-gray-dark-10">
        {t("pages.instances.list.tableHeader.childInstances.label")}
      </div>
      <div className="flex gap-x-1">
        <ConditionalWrapper
          condition={needsPopover}
          wrapper={(children) => (
            <Popover open={popoverOpen} onOpenChange={setPopoverOpen}>
              <PopoverTrigger asChild>{children}</PopoverTrigger>
              <PopoverContent className="w-96 p-0">
                <Command>
                  <CommandInput
                    placeholder={
                      moreInstances
                        ? t(
                            "pages.instances.list.tableHeader.childInstances.searchPlaceholderMax",
                            {
                              count: maxChildInstancesToShow,
                            }
                          )
                        : t(
                            "pages.instances.list.tableHeader.childInstances.searchPlaceholder"
                          )
                    }
                  />
                  <CommandList className="max-h-[278px]">
                    <CommandEmpty>
                      {t(
                        "pages.instances.list.tableHeader.childInstances.notFound"
                      )}
                    </CommandEmpty>
                    <CommandGroup>
                      {data.instances.results.map((instance) => (
                        <CommandItem
                          key={instance.id}
                          value={instance.id}
                          onSelect={(currentValue: string) => {
                            onInstanceSelect(currentValue);
                            setPopoverOpen(false);
                          }}
                        >
                          <div className="flex gap-x-4">
                            <Badge
                              variant={statusToBadgeVariant(instance.status)}
                              className="font-normal"
                              icon={instance.status}
                            />
                            <Badge variant="outline">
                              {instance.id.slice(0, 8)}
                            </Badge>

                            {instance.as}
                          </div>
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>
          )}
        >
          <span
            className={twMergeClsx(needsPopover && "cursor-pointer underline")}
          >
            {moreInstances
              ? t(
                  "pages.instances.list.tableHeader.childInstances.instanceCountMax",
                  {
                    count: maxChildInstancesToShow,
                  }
                )
              : t(
                  "pages.instances.list.tableHeader.childInstances.instanceCount",
                  {
                    count: childCount,
                  }
                )}
          </span>
        </ConditionalWrapper>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger>
              <Button
                icon
                size="sm"
                variant="ghost"
                className="relative -top-0.5"
                disabled={isFetching}
                onClick={() => {
                  refetch();
                }}
              >
                <RefreshCcw />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              {t(
                `pages.instances.list.tableHeader.childInstances.updateTooltip`
              )}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
    </div>
  );
};

export default ChildInstances;
