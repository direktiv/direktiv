import { Check, Circle, Loader2, Square } from "lucide-react";
import {
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandStaticItem,
} from "~/design/Command";

import { twMergeClsx } from "~/util/helpers";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

type NamespaceSelectorListProps = {
  onSelectNamespace: (value: string) => void;
  isMulti?: boolean;
  selectedValues?: string[];
};

export const NamespaceSelectorList = ({
  onSelectNamespace,
  isMulti = false,
  selectedValues = [],
}: NamespaceSelectorListProps) => {
  const { t } = useTranslation();
  const namespace = useNamespace();

  const {
    data: availableNamespaces,
    isLoading,
    isSuccess,
  } = useListNamespaces();

  const hasResults = isSuccess && availableNamespaces?.data.length > 0;

  return (
    <>
      <CommandInput
        placeholder={t("components.breadcrumb.searchPlaceholder")}
      />
      {hasResults && (
        <CommandList className="max-h-[278px]">
          <CommandEmpty>{t("components.breadcrumb.notFound")}</CommandEmpty>
          <CommandGroup>
            {availableNamespaces?.data.map((ns) => (
              <CommandItem
                key={ns.name}
                value={ns.name}
                onSelect={(value) => onSelectNamespace(value)}
              >
                {isMulti ? (
                  <>
                    {selectedValues.includes(ns.name) ? (
                      <Check className="mr-2 h-5 w-5" />
                    ) : (
                      <Square className="mr-2 h-5 w-5" />
                    )}
                    <span>{ns.name}</span>
                  </>
                ) : (
                  <>
                    <Circle
                      className={twMergeClsx(
                        "mr-2 h-2 w-2 fill-current",
                        namespace === ns.name ? "opacity-100" : "opacity-0"
                      )}
                    />
                    <span>{ns.name}</span>
                  </>
                )}
              </CommandItem>
            ))}
          </CommandGroup>
        </CommandList>
      )}
      {isLoading && (
        <CommandStaticItem>
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          {t("components.breadcrumb.loading")}
        </CommandStaticItem>
      )}
    </>
  );
};
