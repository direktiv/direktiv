import { Fragment, useState } from "react";
import {
  Variablepicker,
  VariablepickerError,
  VariablepickerHeading,
  VariablepickerItem,
  VariablepickerMessage,
  VariablepickerSeparator,
} from "~/design/VariablePicker";

import { ButtonBar } from "~/design/ButtonBar";
import Input from "~/design/Input";
import { VarSchema } from "~/api/variables/schema";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";
import { useVars } from "~/api/variables/query/useVariables";
import { z } from "zod";

type variableType = z.infer<typeof VarSchema>;

const NamespaceVariablePicker = ({
  namespace: givenNamespace,
  defaultVariable,
  onChange,
}: {
  namespace?: string;
  defaultVariable?: string;
  onChange: (variable: variableType | undefined) => void;
}) => {
  const { t } = useTranslation();

  const emptyVariable: variableType = {
    name: "",
    checksum: "",
    createdAt: "",
    updatedAt: "",
    size: "",
    mimeType: "",
  };

  const [variableName, setVariableName] = useState(
    defaultVariable ? defaultVariable : ""
  );

  const defaultNamespace = useNamespace();
  const namespace = givenNamespace ? givenNamespace : defaultNamespace;

  const { data, isError } = useVars({ namespace });
  const path = namespace;

  const variableList = data?.variables.results
    ? data.variables.results
    : [emptyVariable];

  const pathNotFound = isError;

  const setNewVariable = (value: string) => {
    emptyVariable.name = value;
    onChange(emptyVariable);
    setVariableName(value);
  };

  const setExistingVariable = (value: string) => {
    const foundVariable = variableList.filter(
      (element: variableType) => element.name === value
    )[0];

    if (foundVariable != undefined) {
      onChange(foundVariable);
      setVariableName(value);
    }
  };

  return (
    <ButtonBar>
      {pathNotFound ? (
        <VariablepickerError
          buttonText={t("components.namespaceVariablepicker.buttonText")}
        >
          <VariablepickerHeading>
            {t("components.namespaceVariablepicker.title", { path })}
          </VariablepickerHeading>
          <VariablepickerSeparator />

          <VariablepickerMessage>
            {t("components.namespaceVariablepicker.error.title", { path })}
          </VariablepickerMessage>
          <VariablepickerSeparator />
        </VariablepickerError>
      ) : (
        <>
          {!variableList || variableList.length == 0 ? (
            <VariablepickerError
              buttonText={t("components.namespaceVariablepicker.buttonText")}
            >
              <VariablepickerHeading>
                {t("components.namespaceVariablepicker.title", { path })}
              </VariablepickerHeading>
              <VariablepickerSeparator />

              <VariablepickerMessage>
                {t("components.namespaceVariablepicker.emptyDirectory.title", {
                  path,
                })}
              </VariablepickerMessage>
              <VariablepickerSeparator />
            </VariablepickerError>
          ) : (
            <Variablepicker
              buttonText={t("components.namespaceVariablepicker.buttonText")}
              value={variableName}
              onValueChange={(value) => {
                setExistingVariable(value);
              }}
            >
              <VariablepickerHeading>
                {t("components.namespaceVariablepicker.title", { path })}
              </VariablepickerHeading>
              <VariablepickerSeparator />
              {variableList.map((variable, index) => (
                <Fragment key={index}>
                  <VariablepickerItem value={variable.name}>
                    {variable.name}
                  </VariablepickerItem>
                  <VariablepickerSeparator />
                </Fragment>
              ))}
            </Variablepicker>
          )}
        </>
      )}
      <Input
        placeholder={t("components.namespaceVariablepicker.placeholder")}
        value={variableName}
        onChange={(e) => {
          setNewVariable(e.target.value);
        }}
      />
    </ButtonBar>
  );
};

export default NamespaceVariablePicker;
