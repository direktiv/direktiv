import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";
import { FC, useState } from "react";
import {
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
} from "~/design/Popover";

import Button from "~/design/Button";
import { Check } from "lucide-react";
import { ContextVariables } from "../../../PageCompiler/primitives/Variable/VariableContext";
import { useTranslation } from "react-i18next";

type VariablePickerProps = {
  variables: ContextVariables;
  onSubmit: (value: string) => void;
  container?: HTMLDivElement;
};

type VariableBuilderState = null | {
  namespace: string;
  id?: string;
  idOptions?: string[];
};

export const VariablePicker: FC<VariablePickerProps> = ({
  variables,
  container,
  onSubmit,
}) => {
  const { t } = useTranslation();

  const [variableBuilder, setVariableBuilder] =
    useState<VariableBuilderState>(null);

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" type="button">
          {t("direktivPage.blockEditor.smartInput.variableBtn")}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="start" container={container}>
        {!variableBuilder?.namespace && (
          <Command>
            <CommandInput placeholder="select variable namespace" />
            <CommandList>
              <CommandGroup heading="namespace">
                {Object.entries(variables).map(([namespace, blockIds]) => (
                  <CommandItem
                    key={namespace}
                    onSelect={() =>
                      setVariableBuilder({
                        namespace,
                        idOptions: Object.keys(blockIds),
                      })
                    }
                  >
                    {namespace}
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        )}
        {!!variableBuilder?.namespace && variableBuilder.idOptions && (
          <Command>
            <CommandInput placeholder="select block id" />
            <CommandList>
              <CommandGroup heading="block scope">
                {variableBuilder.idOptions.map((id) => (
                  <CommandItem
                    key={id}
                    onSelect={() =>
                      setVariableBuilder({
                        ...variableBuilder,
                        id,
                        idOptions: [],
                      })
                    }
                  >
                    {id}
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        )}
        {!!variableBuilder?.namespace && variableBuilder.id && (
          <div>
            <PopoverClose>
              <Button
                variant="outline"
                icon
                type="button"
                onClick={() => {
                  onSubmit(
                    ` {{${variableBuilder.namespace}.${variableBuilder.id}}}`
                  );
                  setVariableBuilder(null);
                }}
              >
                <Check />
              </Button>
            </PopoverClose>
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
};
