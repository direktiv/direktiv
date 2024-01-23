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
import { useWorkflowVariables } from "~/api/tree/query/variables";
import { z } from "zod";

const variableWithoutChecksumAndSize = VarSchema.omit({
  checksum: true,
  size: true,
});

type variableType = z.infer<typeof variableWithoutChecksumAndSize>;

const WorkflowVariablePicker = ({
  namespace: givenNamespace,
  workflow,
  defaultVariable,
  onChange,
}: {
  namespace?: string;
  workflow: string;
  defaultVariable?: variableType;
  onChange: (variable: variableType | undefined) => void;
}) => {
  const { t } = useTranslation();

  const emptyVariable: variableType = {
    name: "",
    createdAt: "",
    updatedAt: "",
    mimeType: "",
  };
  const [inputValue, setInput] = useState(
    defaultVariable ? defaultVariable.name : ""
  );

  const [variable, setVariable] = useState(
    defaultVariable ? defaultVariable : emptyVariable
  );

  const defaultNamespace = useNamespace();
  const namespace = givenNamespace ? givenNamespace : defaultNamespace;

  const path = workflow;

  const { data, isError } = useWorkflowVariables({ path, namespace });

  const variableList = data?.variables.results
    ? data.variables.results
    : [emptyVariable];

  const pathNotFound = isError;

  const setNewVariable = (value: string) => {
    emptyVariable.name = value;
    setVariable(emptyVariable);
    onChange(emptyVariable);
    setInput(value);
  };

  const setExistingVariable = (value: string) => {
    const foundVariable = variableList.filter(
      (element: variableType) => element.name === value
    )[0];

    if (foundVariable != undefined) {
      setVariable(foundVariable);
      onChange(foundVariable);
      setInput(value);
    }
  };

  return (
    <ButtonBar>
      {pathNotFound ? (
        <VariablepickerError
          buttonText={t("components.workflowVariablepicker.buttonText")}
        >
          <VariablepickerHeading>
            {t("components.workflowVariablepicker.title", { path })}
          </VariablepickerHeading>
          <VariablepickerSeparator />

          <VariablepickerMessage>
            {t("components.workflowVariablepicker.error.title", { path })}
          </VariablepickerMessage>
          <VariablepickerSeparator />
        </VariablepickerError>
      ) : (
        <>
          {!variableList || variableList.length == 0 ? (
            <VariablepickerError
              buttonText={t("components.workflowVariablepicker.buttonText")}
            >
              <VariablepickerHeading>
                {t("components.workflowVariablepicker.title", { path })}
              </VariablepickerHeading>
              <VariablepickerSeparator />

              <VariablepickerMessage>
                {t("components.workflowVariablepicker.emptyDirectory.title", {
                  path,
                })}
              </VariablepickerMessage>
              <VariablepickerSeparator />
            </VariablepickerError>
          ) : (
            <Variablepicker
              buttonText={t("components.workflowVariablepicker.buttonText")}
              value={inputValue}
              onValueChange={(value) => {
                setExistingVariable(value);
              }}
            >
              <VariablepickerHeading>
                {t("components.workflowVariablepicker.title", { path })}
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
        placeholder={t("components.workflowVariablepicker.placeholder")}
        value={inputValue}
        onChange={(e) => {
          setNewVariable(e.target.value);
        }}
      />
    </ButtonBar>
  );
};

export default WorkflowVariablePicker;
