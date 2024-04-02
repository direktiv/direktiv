import { Fragment, useState } from "react";
import {
  Variablepicker,
  VariablepickerHeading,
  VariablepickerItem,
  VariablepickerSeparator,
} from "~/design/VariablePicker";

import { ButtonBar } from "~/design/ButtonBar";
import Input from "~/design/Input";
import { VariablePickerError } from "./VariablePickerError";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";
import { useVars } from "~/api/variables_obsolete/query/useVariables";

type NamespaceVariablePickerProps = {
  namespace?: string;
  defaultVariable?: string;
  onChange: (name: string, mimeType?: string) => void;
};

const NamespaceVariablePicker = ({
  namespace: givenNamespace,
  defaultVariable,
  onChange,
}: NamespaceVariablePickerProps) => {
  const { t } = useTranslation();

  const [inputValue, setInput] = useState(defaultVariable ?? "");

  const defaultNamespace = useNamespace();
  const namespace = givenNamespace ?? defaultNamespace;

  const { data, isError: pathNotFound } = useVars({ namespace });

  const variables = data?.variables.results ?? [];
  const noVarsInNamespace = variables.length === 0;

  const setNewVariable = (name: string) => {
    onChange(name);
    setInput(name);
  };

  const setExistingVariable = (name: string) => {
    const foundVariable = variables.find((element) => element.name === name);
    if (foundVariable) {
      onChange(foundVariable?.name, foundVariable?.mimeType);
      setInput(name);
    }
  };

  const getErrorComponent = () => {
    if (pathNotFound) {
      return (
        <VariablePickerError>
          {t("components.namespaceVariablepicker.error.pathNotFound", {
            path: namespace,
          })}
        </VariablePickerError>
      );
    }

    if (noVarsInNamespace) {
      return (
        <VariablePickerError>
          {t("components.namespaceVariablepicker.error.noVarsInNamespace", {
            path: namespace,
          })}
        </VariablePickerError>
      );
    }
    return null;
  };

  const errorComponent = getErrorComponent();

  return (
    <ButtonBar>
      {errorComponent ?? (
        <Variablepicker
          buttonText={t("components.namespaceVariablepicker.buttonText")}
          value={inputValue}
          onValueChange={(value) => {
            setExistingVariable(value);
          }}
        >
          <VariablepickerHeading>
            {t("components.namespaceVariablepicker.title", { path: namespace })}
          </VariablepickerHeading>
          <VariablepickerSeparator />
          {variables.map((variable, index) => (
            <Fragment key={index}>
              <VariablepickerItem value={variable.name}>
                {variable.name}
              </VariablepickerItem>
              <VariablepickerSeparator />
            </Fragment>
          ))}
        </Variablepicker>
      )}

      <Input
        placeholder={t("components.namespaceVariablepicker.placeholder")}
        value={inputValue}
        onChange={(e) => {
          setNewVariable(e.target.value);
        }}
      />
    </ButtonBar>
  );
};

export default NamespaceVariablePicker;
